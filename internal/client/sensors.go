package client

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/Johny2x4/system-telemetry/internal/models"
)

const (
	// Conversion factor for Bytes to Gigabytes
	gb = 1024 * 1024 * 1024
)

// GetRAMMetrics queries the system's virtual and physical memory.
func GetRAMMetrics() (models.RAMMetrics, error) {
	var metrics models.RAMMetrics

	vMem, err := mem.VirtualMemory()
	if err != nil {
		return metrics, fmt.Errorf("failed to retrieve RAM metrics: %w", err)
	}

	// This is especially crucial for tracking bottlenecks when saturating 
	// large memory pools (like 128GB DDR5 setups) during intensive local AI generation.
	metrics.TotalCapacityGB = float64(vMem.Total) / float64(gb)
	metrics.UsedCapacityGB = float64(vMem.Used) / float64(gb)
	metrics.UtilizationPct = vMem.UsedPercent

	return metrics, nil
}

// GetCPUMetrics queries processor utilization, core counts, and temperature.
func GetCPUMetrics() (models.CPUMetrics, error) {
	var metrics models.CPUMetrics

	// Get hardware info (Model Name, Cores)
	info, err := cpu.Info()
	if err != nil {
		return metrics, fmt.Errorf("failed to retrieve CPU info: %w", err)
	}
	if len(info) > 0 {
		metrics.ModelName = info[0].ModelName
	}

	// Logical core count
	cores, err := cpu.Counts(true)
	if err == nil {
		metrics.Cores = cores
	}

	// Global utilization (Sampled over 1 second to get an accurate reading)
	globalPct, err := cpu.Percent(1*time.Second, false)
	if err == nil && len(globalPct) > 0 {
		metrics.GlobalUtil = globalPct[0]
	}

	// Per-core utilization
	perCorePct, err := cpu.Percent(0, true) // 0 duration because the global call just waited 1 second
	if err == nil {
		metrics.PerCoreUtil = perCorePct
	}

	// Attempt to get temperatures (Graceful degradation if unsupported by OS/Permissions)
	metrics.Temperature = getCPUTemperature()

	return metrics, nil
}

// getCPUTemperature attempts to parse sensor data.
// Returns 0.0 if unavailable.
func getCPUTemperature() float64 {
	sensors, err := host.SensorsTemperatures()
	if err != nil || len(sensors) == 0 {
		return 0.0
	}

	// Some systems report multiple thermal zones. We'll grab the first valid CPU/Package temp.
	for _, sensor := range sensors {
		// Sensor keys vary wildly between Windows, Linux, and Mac. 
		// "coretemp" or "k10temp" are common Linux/WSL indicators.
		if sensor.Temperature > 0.0 {
			return sensor.Temperature
		}
	}
	return 0.0
}

// GetDiskMetrics checks capacity and utilization of physical drives.
func GetDiskMetrics() ([]models.DiskMetrics, error) {
	var metrics []models.DiskMetrics

	// Passing 'false' filters out loopback and virtual file systems (like snap/tmpfs)
	partitions, err := disk.Partitions(false)
	if err != nil {
		return metrics, fmt.Errorf("failed to get disk partitions: %w", err)
	}

	for _, p := range partitions {
		// Grab live usage stats for each mount point
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue // Skip if the OS blocks access to a specific drive
		}

		metrics = append(metrics, models.DiskMetrics{
			Device:     p.Device,
			MountPoint: p.Mountpoint,
			FSType:     p.Fstype,
			TotalGB:    float64(usage.Total) / float64(gb),
			UsedGB:     float64(usage.Used) / float64(gb),
			UtilPct:    usage.UsedPercent,
		})
	}
	return metrics, nil
}

// GetNetworkMetrics retrieves IP, MAC, and live IO traffic per interface.
func GetNetworkMetrics() ([]models.NetworkMetrics, error) {
	var metrics []models.NetworkMetrics

	// Get hardware interfaces (IPs, MACs)
	interfaces, err := net.Interfaces()
	if err != nil {
		return metrics, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	// Get IO counters (Total Bytes Up/Down)
	// Passing 'true' gets stats per-interface instead of a global total
	ioCounters, err := net.IOCounters(true)
	if err != nil {
		return metrics, fmt.Errorf("failed to get network IO: %w", err)
	}

	// Map IO counters by interface name for a quick lookup
	ioMap := make(map[string]net.IOCountersStat)
	for _, io := range ioCounters {
		ioMap[io.Name] = io
	}

	for _, intf := range interfaces {
		// Clean up the data: Skip interfaces with no MAC address or no assigned IPs
		if intf.HardwareAddr == "" || len(intf.Addrs) == 0 {
			continue
		}

		var ips []string
		for _, addr := range intf.Addrs {
			ips = append(ips, addr.Addr)
		}

		stat := models.NetworkMetrics{
			InterfaceName: intf.Name,
			IPAddresses:   ips,
			MACAddress:    intf.HardwareAddr,
		}

		// Attach the traffic data if we found it
		if io, exists := ioMap[intf.Name]; exists {
			stat.BytesSent = io.BytesSent
			stat.BytesRecv = io.BytesRecv
		}

		metrics = append(metrics, stat)
	}
	return metrics, nil
}

// CollectAllMetrics gathers all hardware data into a single payload
func CollectAllMetrics(nodeRole string) (models.TelemetryPayload, error) {
	hostInfo, _ := host.Info()

	payload := models.TelemetryPayload{
		Timestamp:  time.Now().UTC(),
		NodeRole:   nodeRole,
		SystemName: hostInfo.Hostname,
		OS:         fmt.Sprintf("%s %s (%s)", hostInfo.OS, hostInfo.PlatformVersion, hostInfo.KernelArch),
	}

	// We won't let one failed sensor crash the whole payload, so we just log errors 
	// internally or ignore them to ensure the server still gets partial data.
	if cpuData, err := GetCPUMetrics(); err == nil {
		payload.CPU = cpuData
	}
	if ramData, err := GetRAMMetrics(); err == nil {
		payload.RAM = ramData
	}
	if diskData, err := GetDiskMetrics(); err == nil {
		payload.Disks = diskData
	}
	if netData, err := GetNetworkMetrics(); err == nil {
		payload.Network = netData
	}
	if gpuData, err := GetGPUMetrics(); err == nil {
		payload.GPUs = gpuData
	} else {
		// If it fails (e.g., this runs on a machine with no NVIDIA card), 
		// it just silently omits the GPU array from the JSON payload.
	}

	return payload, nil
}