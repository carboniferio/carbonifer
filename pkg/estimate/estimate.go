package estimate

import (
	"github.com/carboniferio/carbonifer/internal/estimate"
	internalResources "github.com/carboniferio/carbonifer/internal/resources"
	"github.com/carboniferio/carbonifer/internal/utils"
	"github.com/carboniferio/carbonifer/pkg/providers"
	"github.com/carboniferio/carbonifer/pkg/resources"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
)

type EstimationReport struct {
	Resource        resources.GenericResource
	Power           decimal.Decimal `json:"PowerPerInstance"`
	CarbonEmissions decimal.Decimal `json:"CarbonEmissionsPerInstance"`
	AverageCPUUsage decimal.Decimal
	Count           decimal.Decimal
}

func GetEstimation(resource resources.GenericResource) (EstimationReport, error) {
	estimation, err := estimate.EstimateResource(toInternalComputeResource(resource))
	if err != nil {
		return EstimationReport{}, err
	}
	// Exponent is enforced to avoid equality issues when comparing reports.
	// Indeed, if we don't truncate the values, we might have value with various exponent,
	// which will make the equality check fail during test.
	// TODO: Find a better way to handle this
	return EstimationReport{
		Resource:        resource,
		Power:           estimation.Power.Truncate(10),
		CarbonEmissions: estimation.CarbonEmissions.Truncate(10),
		AverageCPUUsage: estimation.AverageCPUUsage.Truncate(10),
		Count:           estimation.Count.Truncate(10),
	}, nil
}

func GetEstimationFromInstanceType(instanceType string, zone string, provider providers.Provider) (EstimationReport, error) {
	resource, err := resources.GetResource(instanceType, zone, provider)
	if err != nil {
		return EstimationReport{}, err
	}
	estimation, err := GetEstimation(resource)
	if err != nil {
		return EstimationReport{}, err
	}
	return estimation, nil
}

func toInternalComputeResource(resource resources.GenericResource) internalResources.ComputeResource {
	// TODO: Support multiple CPU types
	// Check this PR for more info: https://github.com/carboniferio/carbonifer/pull/41
	return internalResources.ComputeResource{
		Identification: resource.GetIdentification(),
		Specs: &internalResources.ComputeResourceSpecs{
			GpuTypes:          resource.GPUTypes,
			HddStorage:        resource.Storage.HddStorage,
			SsdStorage:        resource.Storage.SsdStorage,
			MemoryMb:          resource.MemoryMb,
			VCPUs:             resource.VCPUs,
			ReplicationFactor: resource.ReplicationFactor,
		},
	}
}

func init() {
	viper.Set("data.path", "../../data")
	utils.InitConfig("../../internal/utils/defaults.yaml")
}
