package models

import "time"

type TelemetryPayload struct {
	Timestamp   time.Time        `json:"timestamp"`
	NodeRole    string           `json:"node_role"` 
	SystemName  string           `json:"system_name"`
	OS          string           `json:"os"`
	Network     []NetworkMetrics `json:"network"`
	CPU         CPUMetrics       `json:"cpu"`
	RAM         RAMMetrics       `json:"ram"`
	Disks       []DiskMetrics    `json:"disks"`
	GPUs       []GPUMetrics     `json:"gpus"`
	// GPU will be added later when we tackle NVML bindings
}

type NetworkMetrics struct {
	InterfaceName string   `json:"interface_name"`
	IPAddresses   []string `json:"ip_addresses"`
	MACAddress    string   `json:"mac_address"`
	BytesSent     uint64   `json:"bytes_sent_sec"`
	BytesRecv     uint64   `json:"bytes_recv_sec"`
}

type CPUMetrics struct {
	ModelName   string    `json:"model_name"`
	Cores       int       `json:"cores"`
	GlobalUtil  float64   `json:"global_utilization"`
	PerCoreUtil []float64 `json:"per_core_utilization"`
	Temperature float64   `json:"temperature_c"` 
}

type RAMMetrics struct {
	TotalCapacityGB float64 `json:"total_capacity_gb"`
	UsedCapacityGB  float64 `json:"used_capacity_gb"`
	UtilizationPct  float64 `json:"utilization_pct"`
}

type DiskMetrics struct {
	Device     string  `json:"device"`
	MountPoint string  `json:"mount_point"`
	FSType     string  `json:"fs_type"`
	TotalGB    float64 `json:"total_gb"`
	UsedGB     float64 `json:"used_gb"`
	UtilPct    float64 `json:"util_pct"`
}

type GPUMetrics struct {
	Name        string  `json:"name"`
	MemoryTotal float64 `json:"memory_total_gb"`
	MemoryUsed  float64 `json:"memory_used_gb"`
	MemoryUtil  float64 `json:"memory_util_pct"`
	ComputeUtil float64 `json:"compute_util_pct"`
	Temperature float64 `json:"temperature_c"`
}