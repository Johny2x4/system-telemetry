package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kardianos/service"
	"github.com/Johny2x4/system-telemetry/internal/client"
	"github.com/Johny2x4/system-telemetry/internal/config"
	"github.com/Johny2x4/system-telemetry/internal/server"
)

// program represents our background service
type program struct {
	exit chan struct{}
	cfg  *config.AppConfig
}

func (p *program) Start(s service.Service) error {
	p.exit = make(chan struct{})
	// Start the actual logic in a background goroutine so the Service Manager doesn't block
	go p.run()
	return nil
}

func (p *program) run() {
	if p.cfg.Role == "Server" {
		log.Println("Starting in SERVER mode...")

		// Initialize SQLite
		db, err := server.NewSQLiteConnector(p.cfg.DBPath)
		if err != nil {
			log.Fatalf("Failed to start DB: %v", err)
		}
		defer db.Close()

		// Start the HTTP API Server (This blocks forever)
		server.StartAPIServer(p.cfg.ListenPort, db)

	} else {
		log.Println("Starting in CLIENT mode...")
		
		// The polling loop
		ticker := time.NewTicker(10 * time.Second) // Poll every 10 seconds
		for {
			select {
			case <-ticker.C:
				payload, err := client.CollectAllMetrics("Client")
				if err != nil {
					log.Printf("Metric collection error: %v", err)
					continue
				}

				// Send the payload to the server
				jsonData, _ := json.Marshal(payload)
				resp, err := http.Post(p.cfg.ServerURL+"/api/v1/ingest", "application/json", bytes.NewBuffer(jsonData))
				
				if err != nil {
					log.Printf("Failed to send data to server: %v", err)
				} else {
					resp.Body.Close()
					log.Println("Payload successfully pushed to server.")
				}

			case <-p.exit:
				ticker.Stop()
				return
			}
		}
	}
}

func (p *program) Stop(s service.Service) error {
	close(p.exit)
	return nil
}

func main() {
	// 1. Load the config (or run the setup wizard if it's the first run)
	cfg, err := config.LoadOrSetup()
	if err != nil {
		log.Fatalf("Setup failed: %v", err)
	}

	// 2. Configure the Service Manager
	svcConfig := &service.Config{
		Name:        "SystemTelemetry",
		DisplayName: "System Telemetry Agent",
		Description: "Collects hardware metrics and reports them to a central dashboard.",
	}

	prg := &program{cfg: cfg}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	// 3. Handle Service Installation/Control commands if passed via CLI
	if len(os.Args) > 1 {
		err = service.Control(s, os.Args[1])
		if err != nil {
			log.Fatalf("Valid actions: %q\n", service.ControlAction)
		}
		return
	}

	// 4. If no CLI arguments were passed, run the program normally
	err = s.Run()
	if err != nil {
		log.Fatal(err)
	}
}