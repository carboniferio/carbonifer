package estimate

import (
	"github.com/carboniferio/carbonifer/internal/estimate"
	internalResources "github.com/carboniferio/carbonifer/internal/resources"
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
	return EstimationReport{
		Resource:        resource,
		Power:           estimation.Power,
		CarbonEmissions: estimation.CarbonEmissions,
		AverageCPUUsage: estimation.AverageCPUUsage,
		Count:           estimation.Count,
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
	return internalResources.ComputeResource{
		Identification: resource.GetIdentification(),
		Specs: &internalResources.ComputeResourceSpecs{
			GpuTypes:          resource.GPUTypes,
			HddStorage:        resource.Storage.HddStorage,
			SsdStorage:        resource.Storage.SsdStorage,
			MemoryMb:          resource.MemoryMb,
			VCPUs:             resource.VCPUs,
			CPUType:           resource.CPUTypes[0], // TODO: Support multiple CPU types
			ReplicationFactor: resource.ReplicationFactor,
		},
	}
}

func init() {
	viper.Set("data.path", "../../data")
	viper.Set("unit.power", "")
	viper.Set("unit.time", "")
	viper.Set("unit.carbon", "")
}
