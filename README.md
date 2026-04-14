
# 🚀 System Telemetry & Monitoring Agent

A high-performance, zero-dependency telemetry suite built in Go. This tool aggregates hardware metrics from across your fleet into a centralized SQLite database, optimized for real-time Grafana dashboards.

## ✨ Key Features
* **Zero-Config Database:** Uses an embedded SQLite backend (`modernc.org/sqlite`). No external database installation or Docker containers required.
* **Interactive Setup:** A built-in CLI wizard configures the node as a **Client** or **Server** on the first run.
* **Daemonization:** Natively installs as a background service on **Windows (Services)**, **Linux (systemd)**, and **macOS (launchd)**.
* **Smart Self-Monitoring:** Server nodes automatically monitor their own hardware while simultaneously ingesting data from remote clients.
* **High-Resolution Polling:** Supports intervals as low as **5 seconds** for tracking intensive AI/rendering workloads.
* **Automatic Data Retention:** Includes a background vacuum that prunes data older than 30 days to keep the database lean.

---

## 🛠️ Installation & Build

### 1. Compile for Windows
```powershell
$env:GOOS="windows"; $env:GOARCH="amd64"
go build -o telemetry.exe ./cmd/telemetry/main.go
```

### 2. Compile for Linux (With NVIDIA GPU Tracking)
*Must be compiled on a Linux machine or WSL2 to enable NVML bindings.*
```bash
go build -o telemetry ./cmd/telemetry/main.go
```

---

## ⚙️ Usage

### Initialization
Simply run the executable. The **Setup Wizard** will guide you through role assignment and networking.
```powershell
./telemetry.exe
```

### Managing the Background Service
Run these commands from an **Administrator/Sudo** terminal to manage the daemon:
```powershell
./telemetry.exe install    # Register with the OS service manager
./telemetry.exe start      # Start background collection
./telemetry.exe stop       # Stop the service
./telemetry.exe uninstall  # Remove from the OS
```

### Reconfiguring
Need to change the polling interval or the Server IP? Use the reconfiguration flag:
```powershell
./telemetry.exe reconfigure
```

---

## 🔒 Networking & Firewalls
For Clients to reach the Server, ensure the Server has an inbound rule for your chosen port (default `8080`).

**Windows Server Firewall Rule:**
```powershell
New-NetFirewallRule -DisplayName "Telemetry" -Direction Inbound -LocalPort 8080 -Protocol TCP -Action Allow -Profile Any
```
