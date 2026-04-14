
## 📝 Change Notes: The "Fleet Evolution" Update

### 🏗️ Architectural Pivots
* **Database Migration:** Swapped InfluxDB for a **Pure-Go SQLite** backend (`modernc.org/sqlite`). This removed the requirement for Docker or external DB installations, enabling a zero-dependency, single-binary deployment.
* **Automated Data Retention:** Implemented a background "vacuum" Goroutine that automatically prunes telemetry records older than 30 days.
* **Daemonization:** Integrated the `kardianos/service` library, allowing the executable to install/uninstall itself as a native background service on Windows, Linux, and macOS.

### ⚙️ CLI & Configuration Improvements
* **Setup Wizard:** Added an interactive CLI wizard using the `survey` library to handle role assignment (Client vs. Server) on the first run.
* **Reconfiguration Support:** Added a `--reconfigure` flag to allow users to update polling intervals or server URLs without manually editing YAML files.
* **URL Sanitization:** Implemented automatic logic to prepend `http://` and strip trailing slashes, preventing common connection string errors.
* **Dynamic Polling:** Added support for user-defined polling intervals, with a safety floor of 5 seconds to ensure high-resolution monitoring without system strain.

### 🛠️ Reliability & Deployment Fixes
* **Working Directory Enforcement:** Added logic to force the application to its own executable directory, fixing a critical bug where Windows Services failed to find configuration files.
* **Server Self-Monitoring:** Updated the Server role to monitor and report its own hardware metrics while simultaneously acting as an API aggregator.
* **Firewall & Network Binding:** Optimized network listeners to bind to all interfaces and documented the necessary universal firewall rules for fleet connectivity.

### 📊 Dashboard & Query Logic
* **Case-Sensitivity Fixes:** Standardized JSON tagging and SQL `json_extract` paths to ensure consistent data flow between Go and SQLite.
* **Advanced SQL Strategy:** Developed a "Latest Per Host" query using Window Functions (`ROW_NUMBER() OVER`) to provide real-time fleet status snapshots in Grafana.
