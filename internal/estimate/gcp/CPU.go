package gcp

import (
	"strings"

	"github.com/carboniferio/carbonifer/internal/estimate/coefficients"
	"github.com/carboniferio/carbonifer/internal/providers/gcp"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
)

func estimateWattGCPCPU(resource *resources.ComputeResource) decimal.Decimal {
	// Get average CPU usage
	averageCPUUse := decimal.NewFromFloat(viper.GetFloat64("provider.gcp.avg_cpu_use"))

	var avgWatts decimal.Decimal
	// Average Watts = Min Watts + Avg vCPU Utilization * (Max Watts - Min Watts)
	cpu_platform := resource.Specs.CPUType
	if cpu_platform != "" {
		cpu_platform := gcp.GetCPUWatt(strings.ToLower(cpu_platform))
		avgWatts = cpu_platform.MinWatts.Add(averageCPUUse.Mul(cpu_platform.MaxWatts.Sub(cpu_platform.MinWatts)))
	} else {
		minWH := coefficients.GetEnergyCoefficients().GCP.CPUMinWh
		maxWh := coefficients.GetEnergyCoefficients().GCP.CPUMaxWh
		avgWatts = minWH.Add(averageCPUUse.Mul(maxWh.Sub(minWH)))
	}
	return avgWatts.Mul(decimal.NewFromInt32(resource.Specs.VCPUs))
}
