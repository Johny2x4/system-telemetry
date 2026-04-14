//go:build linux

package client

import (
	"fmt"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/Johny2x4/system-telemetry/internal/models" 
)

// GetGPUMetrics interfaces directly with the NVIDIA driver to pull hardware states.
func GetGPUMetrics() ([]models.GPUMetrics, error) {
	var metrics []models.GPUMetrics

	// 1. Initialize the NVML Library
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		return metrics, fmt.Errorf("NVML init failed (is an NVIDIA driver installed?): %v", nvml.ErrorString(ret))
	}
	// Important: We must release the NVML handle when this function exits to prevent memory leaks
	defer nvml.Shutdown()

	// 2. Count the number of GPUs attached to the system
	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return metrics, fmt.Errorf("failed to get GPU count: %v", nvml.ErrorString(ret))
	}

	// 3. Loop through all found GPUs and extract telemetry
	for i := 0; i < count; i++ {
		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			continue // Skip this device if the handle fails
		}

		name, _ := device.GetName()

		// Extract VRAM utilization (Critical for tracking model loading bottlenecks)
		memory, ret := device.GetMemoryInfo()
		var memTotal, memUsed, memUtil float64
		if ret == nvml.SUCCESS {
			memTotal = float64(memory.Total) / float64(gb)
			memUsed = float64(memory.Used) / float64(gb)
			if memory.Total > 0 {
				memUtil = (float64(memory.Used) / float64(memory.Total)) * 100.0
			}
		}

		// Extract Core Compute Utilization
		utilization, ret := device.GetUtilizationRates()
		var computeUtil float64
		if ret == nvml.SUCCESS {
			computeUtil = float64(utilization.Gpu)
		}

		// Extract Core Temperature
		temp, ret := device.GetTemperature(nvml.TEMPERATURE_GPU)
		var tempC float64
		if ret == nvml.SUCCESS {
			tempC = float64(temp)
		}

		metrics = append(metrics, models.GPUMetrics{
			Name:        name,
			MemoryTotal: memTotal,
			MemoryUsed:  memUsed,
			MemoryUtil:  memUtil,
			ComputeUtil: computeUtil,
			Temperature: tempC,
		})
	}

	return metrics, nil
}