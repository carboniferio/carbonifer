package estimate

import (
	"fmt"
	"strings"

	"github.com/carboniferio/carbonifer/internal/estimate/coefficients"
	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/providers/gcp"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
)

func estimateWattCPU(resource *resources.ComputeResource) decimal.Decimal {
	provider := resource.Identification.Provider
	// Get average CPU usage
	averageCPUUse := decimal.NewFromFloat(viper.GetFloat64(fmt.Sprintf("provider.%s.avg_cpu_use", provider.String())))

	var avgWatts decimal.Decimal
	// Average Watts = Min Watts + Avg vCPU Utilization * (Max Watts - Min Watts)
	cpuPlatform := resource.Specs.CPUType
	if cpuPlatform != "" && resource.Identification.Provider == providers.GCP {
		cpuPlatform := gcp.GetCPUWatt(strings.ToLower(cpuPlatform))
		avgWatts = cpuPlatform.MinWatts.Add(averageCPUUse.Mul(cpuPlatform.MaxWatts.Sub(cpuPlatform.MinWatts)))
	} else {
		minWH := coefficients.GetEnergyCoefficients().GetByProvider(provider).CPUMinWh
		maxWh := coefficients.GetEnergyCoefficients().GetByProvider(provider).CPUMaxWh
		avgWatts = minWH.Add(averageCPUUse.Mul(maxWh.Sub(minWH)))
	}
	return avgWatts.Mul(decimal.NewFromInt32(resource.Specs.VCPUs))
}
