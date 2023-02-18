package allprovider

import (
	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
)

func EstimateWattGPU(resource *resources.ComputeResource) decimal.Decimal {
	// Get average GPU usage
	averageCPUUse := decimal.NewFromFloat(viper.GetFloat64("provider.gcp.avg_gpu_use"))

	avgWattsTotal := decimal.Zero
	// Average Watts = Min Watts + Avg GPU Utilization * (Max Watts - Min Watts)
	for _, gpuType := range resource.Specs.GpuTypes {
		gpuWatt := providers.GetGPUWatt(gpuType)
		avgWatts := gpuWatt.MinWatts.Add(averageCPUUse.Mul(gpuWatt.MaxWatts.Sub(gpuWatt.MinWatts)))
		avgWattsTotal = avgWattsTotal.Add(avgWatts)
	}
	return avgWattsTotal
}
