# System Telemetry & Monitoring Agent

A cross-platform, zero-dependency telemetry service designed to collect deep system hardware metrics and aggregate them into a centralized SQLite database for Grafana visualization. 

Built entirely in Go, this tool compiles down into a single executable. It features an interactive CLI setup wizard, automatic background service daemonization, and specialized tracking for dedicated GPU VRAM and compute utilization—making it ideal for monitoring physical automation nodes and local AI model hosting environments.

## ✨ Key Features
* **Dual-Mode Architecture:** Run the executable as a lightweight Client Agent to poll hardware, or as a Server Aggregator to ingest data via a REST API.
* **Zero-Setup Database:** The Server mode automatically initializes a local SQLite database (`telemetry.db`) with a background task that prunes data older than 30 days. No Docker or external DB engine required.
* **NVIDIA NVML Integration:** Natively interfaces with NVIDIA drivers to extract highly accurate VRAM saturation, core compute utilization, and thermal metrics.
* **Cross-Platform Daemonization:** Automatically installs itself into the host's native service manager (Windows Services, Linux `systemd`, or macOS `launchd`) to run silently on boot.
* **Grafana Ready:** The SQLite database file acts as a direct, plug-and-play data source for Grafana dashboards.

## 📊 Metrics Collected
* **Host Identity:** OS details, Hostname, IPv4/IPv6 Addresses, and MAC Addresses.
* **Compute:** CPU Model, Global/Per-Core Utilization (%), and Core Temperatures.
* **Memory:** Total RAM Capacity, Used Capacity (GB), and Utilization (%).
* **Storage:** Physical Drive Mount Points, File System Types, and Capacity/Utilization per drive.
* **Network I/O:** Live Bytes Received/Transmitted per active interface.
* **Accelerators (GPU):** Hardware Name, Dedicated VRAM Total/Used (GB), VRAM Utilization (%), Core Compute Utilization (%), and Core Temperatures.

---

## 🚀 Installation & Compilation

Because this project uses conditional build tags to handle CGO requirements for the NVIDIA driver, you must compile the binary specifically for your target environment.

### 1. Windows (Client or Server)
Compile directly from PowerShell. This build will utilize WMI/standard OS polling.
```powershell
go build -o telemetry.exe ./cmd/telemetry/main.go
```

### 2. Linux (Basic Server / No GPU)
If deploying to a standard Linux server that simply needs to aggregate data or monitor basic OS stats, cross-compile from Windows:
```powershell
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o telemetry-linux ./cmd/telemetry/main.go
```

### 3. Linux (AI/GPU Node)
To enable the NVIDIA NVML bindings for deep GPU tracking, you **must** compile from within a Linux environment (like WSL2 or directly on the target machine).
```bash
go build -o telemetry-gpu ./cmd/telemetry/main.go
```

---

## ⚙️ Usage & Configuration

### Step 1: Initialization
Execute the compiled binary for the first time. It will detect that no configuration exists and launch the interactive setup wizard.
```bash
./telemetry
```
* **Server Mode:** It will prompt you for an API listening port and a database filename.
* **Client Mode:** It will prompt you for the HTTP URL of your Server Aggregator (e.g., `http://192.168.1.50:8080`).

### Step 2: Background Service Installation
Once configured, use the built-in CLI commands to install the application into your operating system's background service manager. Note: Run your terminal as an Administrator / `sudo`.

```bash
# Install the background service
./telemetry install

# Start the service
./telemetry start
```
*To stop or uninstall, simply replace `start` with `stop` or `uninstall`.*
