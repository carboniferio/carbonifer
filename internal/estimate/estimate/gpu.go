package estimate

import (
	"fmt"
	"strings"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
)

// EstimateWattGPU estimates the power consumption of a GPU resource
func EstimateWattGPU(resource *resources.ComputeResource) decimal.Decimal {
	// Get average GPU usage
	provider := strings.ToLower(resource.Identification.Provider.String())
	averageCPUUse := decimal.NewFromFloat(viper.GetFloat64(fmt.Sprintf("provider.%s.avg_gpu_use", provider)))

	avgWattsTotal := decimal.Zero
	// Average Watts = Min Watts + Avg GPU Utilization * (Max Watts - Min Watts)
	for _, gpuType := range resource.Specs.GpuTypes {
		gpuWatt := providers.GetGPUWatt(gpuType)
		avgWatts := gpuWatt.MinWatts.Add(averageCPUUse.Mul(gpuWatt.MaxWatts.Sub(gpuWatt.MinWatts)))
		avgWattsTotal = avgWattsTotal.Add(avgWatts)
	}
	return avgWattsTotal
}
