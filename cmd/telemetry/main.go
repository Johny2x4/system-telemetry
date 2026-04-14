package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings" // Added for URL sanitization
	"time"
	"path/filepath"

	"github.com/kardianos/service"
	// IMPORTANT: Ensure these match your go.mod module name
	"github.com/Johny2x4/system-telemetry/internal/client"
	"github.com/Johny2x4/system-telemetry/internal/config"
	"github.com/Johny2x4/system-telemetry/internal/server"
)

type program struct {
	exit chan struct{}
	cfg  *config.AppConfig
}

func (p *program) Start(s service.Service) error {
	p.exit = make(chan struct{})
	go p.run()
	return nil
}

func (p *program) run() {
	interval := p.cfg.PollingInterval
	if interval < 5 {
		interval = 5
	}

	if p.cfg.Role == "Server" {
		log.Printf("Starting in SERVER mode on port %s...", p.cfg.ListenPort)

		db, err := server.NewSQLiteConnector(p.cfg.DBPath)
		if err != nil {
			log.Fatalf("Critical: Failed to initialize database: %v", err)
		}
		defer db.Close()

		go func() {
			ticker := time.NewTicker(time.Duration(interval) * time.Second)
			for {
				select {
				case <-ticker.C:
					payload, err := client.CollectAllMetrics("Server")
					if err == nil {
						_ = db.WritePayload(payload)
					}
				case <-p.exit:
					ticker.Stop()
					return
				}
			}
		}()

		server.StartAPIServer(p.cfg.ListenPort, db)

	} else {
		// --- CLIENT LOGIC ---
		
		// Sanitize the URL: ensure it has http:// and no trailing slashes
		cleanURL := p.cfg.ServerURL
		if !strings.HasPrefix(cleanURL, "http://") && !strings.HasPrefix(cleanURL, "https://") {
			cleanURL = "http://" + cleanURL
		}
		cleanURL = strings.TrimSuffix(cleanURL, "/")
		
		targetURL := fmt.Sprintf("%s/api/v1/ingest", cleanURL)
		
		log.Printf("Starting in CLIENT mode. Target: %s", targetURL)
		
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		clientHTTP := &http.Client{Timeout: 5 * time.Second}

		for {
			select {
			case <-ticker.C:
				payload, err := client.CollectAllMetrics("Client")
				if err != nil {
					log.Printf("Metric collection error: %v", err)
					continue
				}

				jsonData, _ := json.Marshal(payload)
				
				// We use the 'targetURL' variable we defined above
				resp, err := clientHTTP.Post(targetURL, "application/json", bytes.NewBuffer(jsonData))
				if err != nil {
					log.Printf("Network Error: %v", err)
					continue
				}
				resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					log.Printf("Server Error: Status %d", resp.StatusCode)
				} else {
					log.Println("Telemetry successfully pushed to server.")
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
	execPath, err := os.Executable()
    if err == nil {
        os.Chdir(filepath.Dir(execPath))
    }
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "reconfigure" || arg == "--reconfigure" || arg == "setup" {
			_, err := config.RunSetupWizard()
			if err != nil {
				log.Fatalf("Reconfiguration failed: %v", err)
			}
			fmt.Println("Reconfiguration successful. Restart the service to apply.")
			return
		}
	}

	cfg, err := config.LoadOrSetup()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	svcConfig := &service.Config{
		Name:        "SystemTelemetry",
		DisplayName: "System Telemetry Agent",
		Description: "Unified hardware monitoring agent.",
	}

	prg := &program{cfg: cfg}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 1 {
		cmd := os.Args[1]
		err = service.Control(s, cmd)
		if err != nil {
			log.Fatalf("Failed to execute '%s': %v.", cmd, err)
		}
		fmt.Printf("Service command '%s' executed successfully.\n", cmd)
		return
	}

	err = s.Run()
	if err != nil {
		log.Fatal(err)
	}
}