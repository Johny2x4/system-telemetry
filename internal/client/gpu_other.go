//go:build !linux

package client

import (
	"github.com/Johny2x4/system-telemetry/internal/models"
)

// GetGPUMetrics acts as a safe stub for non-Linux systems.
// (We can always integrate Windows WMI GPU polling here later!)
func GetGPUMetrics() ([]models.GPUMetrics, error) {
	var metrics []models.GPUMetrics
	// Return an empty array so the JSON payload simply omits GPUs
	// without crashing the client agent.
	return metrics, nil
}