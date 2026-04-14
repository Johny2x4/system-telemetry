package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"

	"github.com/Johny2x4/system-telemetry/internal/models"
)

type SQLiteConnector struct {
	db *sql.DB
}

// NewSQLiteConnector initializes the local database file and tables
func NewSQLiteConnector(dbPath string) (*SQLiteConnector, error) {
	// Connect to the SQLite file (creates it if it doesn't exist)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite db: %w", err)
	}

	// Create our table. We extract the most common metrics into their own columns 
	// for easy Grafana graphing, and store the rest of the arrays (disks, gpus) in a JSON column.
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS telemetry (
		timestamp DATETIME,
		node_role TEXT,
		system_name TEXT,
		cpu_util REAL,
		cpu_temp REAL,
		ram_util REAL,
		ram_used_gb REAL,
		full_payload JSON
	);
	CREATE INDEX IF NOT EXISTS idx_timestamp ON telemetry(timestamp);
	CREATE INDEX IF NOT EXISTS idx_system_name ON telemetry(system_name);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	connector := &SQLiteConnector{db: db}

	// Start the automated 30-day data retention background vacuum
	go connector.startCleanupTask()

	return connector, nil
}

// WritePayload inserts the telemetry data into the local database
func (sq *SQLiteConnector) WritePayload(payload models.TelemetryPayload) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload to json: %w", err)
	}

	insertSQL := `
	INSERT INTO telemetry (timestamp, node_role, system_name, cpu_util, cpu_temp, ram_util, ram_used_gb, full_payload) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = sq.db.Exec(insertSQL,
		payload.Timestamp.Format(time.RFC3339),
		payload.NodeRole,
		payload.SystemName,
		payload.CPU.GlobalUtil,
		payload.CPU.Temperature,
		payload.RAM.UtilizationPct,
		payload.RAM.UsedCapacityGB,
		string(payloadJSON),
	)

	return err
}

// startCleanupTask runs once every 24 hours to delete data older than 30 days
func (sq *SQLiteConnector) startCleanupTask() {
	ticker := time.NewTicker(24 * time.Hour)
	for range ticker.C {
		log.Println("Running automated 30-day database cleanup...")
		cleanupSQL := `DELETE FROM telemetry WHERE timestamp < datetime('now', '-30 days');`
		
		res, err := sq.db.Exec(cleanupSQL)
		if err != nil {
			log.Printf("Database cleanup failed: %v", err)
			continue
		}
		
		rows, _ := res.RowsAffected()
		log.Printf("Cleanup complete. Removed %d old records.", rows)
	}
}

func (sq *SQLiteConnector) Close() {
	sq.db.Close()
}