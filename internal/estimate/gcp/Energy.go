package gcp

import (
	"github.com/carboniferio/carbonifer/internal/estimate/allprovider"
	"github.com/carboniferio/carbonifer/internal/estimate/coefficients"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

// Source: https://www.cloudcarbonfootprint.org/docs/methodology/#appendix-i-energy-coefficients
// in Watt Hour
func EstimateWattHourGCP(resource *resources.ComputeResource) decimal.Decimal {
	cpuEstimationInWh := estimateWattGCPCPU(resource)
	log.Debugf("%v.%v CPU in Wh: %v", resource.Identification.ResourceType, resource.Identification.Name, cpuEstimationInWh)
	memoryEstimationInWH := estimateWattMem(resource)
	log.Debugf("%v.%v Memory in Wh: %v", resource.Identification.ResourceType, resource.Identification.Name, memoryEstimationInWH)
	storageInWh := estimateWattStorage(resource)
	log.Debugf("%v.%v Storage in Wh: %v", resource.Identification.ResourceType, resource.Identification.Name, storageInWh)
	gpuEstimationInWh := allprovider.EstimateWattGPU(resource)
	log.Debugf("%v.%v GPUs in Wh: %v", resource.Identification.ResourceType, resource.Identification.Name, gpuEstimationInWh)
	pue := coefficients.GetEnergyCoefficients().GCP.PueAverage
	log.Debugf("%v.%v PUE %v", resource.Identification.ResourceType, resource.Identification.Name, pue)
	rawWattEstimate := decimal.Sum(
		cpuEstimationInWh,
		memoryEstimationInWH,
		storageInWh,
		gpuEstimationInWh,
	)
	replicationFactor := resource.Specs.ReplicationFactor
	if replicationFactor == 0 {
		replicationFactor = 1
	}
	wattEstimate := pue.Mul(rawWattEstimate).Mul(decimal.NewFromInt32(replicationFactor))
	log.Debugf("%v.%v Energy in Wh: %v", resource.Identification.ResourceType, resource.Identification.Name, wattEstimate)
	return wattEstimate
}
