package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Johny2x4/system-telemetry/internal/models"
)

// StartAPIServer boots the REST API and attaches the SQLite database
func StartAPIServer(port string, db *SQLiteConnector) {
	http.HandleFunc("/api/v1/ingest", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload models.TelemetryPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			log.Printf("Failed to decode payload: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Push the decoded data into the local SQLite file
		if err := db.WritePayload(payload); err != nil {
			log.Printf("DB Write Error: %v", err)
			http.Error(w, "Internal database error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Telemetry ingested successfully")
	})

	log.Printf("Starting Server Node on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server crashed: %v", err)
	}
}