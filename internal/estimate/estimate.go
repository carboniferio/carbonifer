package estimate

import (
	"fmt"
	"time"

	"github.com/carboniferio/carbonifer/internal/estimate/estimate"
	"github.com/carboniferio/carbonifer/internal/estimate/estimation"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func EstimateResources(resourceList map[string]resources.Resource) estimation.EstimationReport {

	var estimationResources []estimation.EstimationResource
	var unsupportedResources []resources.Resource
	estimationTotal := estimation.EstimationTotal{
		Power:           decimal.Zero,
		CarbonEmissions: decimal.Zero,
		ResourcesCount:  decimal.Zero,
	}
	for _, resource := range resourceList {
		estimationResource, uerr := EstimateResource(resource)
		if uerr != nil {
			logrus.Warnf("Skipping unsupported provider %v: %v.%v", uerr.Provider, resource.GetIdentification().ResourceType, resource.GetIdentification().Name)
		}

		if resource.IsSupported() {
			estimationResources = append(estimationResources, *estimationResource)
		} else {
			unsupportedResources = append(unsupportedResources, resource)
		}

		estimationTotal.Power = estimationTotal.Power.Add(estimationResource.Power.Mul(estimationResource.Count))
		estimationTotal.CarbonEmissions = estimationTotal.CarbonEmissions.Add(estimationResource.CarbonEmissions.Mul(estimationResource.Count))
		estimationTotal.ResourcesCount = estimationTotal.ResourcesCount.Add(estimationResource.Count)
	}

	return estimation.EstimationReport{
		Info: estimation.EstimationInfo{
			UnitTime:                viper.Get("unit.time").(string),
			UnitWattTime:            fmt.Sprintf("%s%s", viper.Get("unit.power"), viper.Get("unit.time")),
			UnitCarbonEmissionsTime: fmt.Sprintf("%sCO2eq/%s", viper.Get("unit.carbon"), viper.Get("unit.time")),
			DateTime:                time.Now(),
			InfoByProvider: map[providers.Provider]estimation.InfoByProvider{
				providers.GCP: {
					AverageCPUUsage: viper.GetFloat64("provider.gcp.avg_cpu_use"),
					AverageGPUUsage: viper.GetFloat64("provider.gcp.avg_gpu_use"),
				},
				providers.AWS: {
					AverageCPUUsage: viper.GetFloat64("provider.gcp.avg_cpu_use"),
					AverageGPUUsage: viper.GetFloat64("provider.gcp.avg_gpu_use"),
				},
			},
		},
		Resources:            estimationResources,
		UnsupportedResources: unsupportedResources,
		Total:                estimationTotal,
	}

}

func EstimateResource(resource resources.Resource) (*estimation.EstimationResource, *providers.UnsupportedProviderError) {
	if !resource.IsSupported() {
		return estimateNotSupported(resource.(resources.UnsupportedResource)), nil
	}
	switch resource.GetIdentification().Provider {
	case providers.AWS:
		return estimate.EstimateSupportedResource(resource), nil
	case providers.GCP:
		return estimate.EstimateSupportedResource(resource), nil
	default:
		return nil, &providers.UnsupportedProviderError{Provider: resource.GetIdentification().Provider.String()}
	}
}

func estimateNotSupported(resource resources.UnsupportedResource) *estimation.EstimationResource {
	return &estimation.EstimationResource{
		Resource:        resource,
		Power:           decimal.Zero,
		CarbonEmissions: decimal.Zero,
		AverageCPUUsage: decimal.Zero,
	}
}
