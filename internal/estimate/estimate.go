package estimate

import (
	"fmt"
	"time"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func EstimateResources(resources []resources.ComputeResource) EstimationReport {

	var estimationResources []EstimationResource
	estimationTotal := EstimationTotal{
		Power:           decimal.Zero,
		CarbonEmissions: decimal.Zero,
		ResourcesCount:  0,
	}
	for _, resource := range resources {
		estimationResource, uerr := EstimateResource(resource)
		if uerr != nil {
			logrus.Warnf("Skipping unsupported provider %v: %v.%v", uerr.Provider, resource.ResourceType, resource.Name)
		}

		estimationResources = append(estimationResources, *estimationResource)
		estimationTotal.Power = estimationTotal.Power.Add(estimationResource.CarbonEmissions)
		estimationTotal.CarbonEmissions = estimationTotal.CarbonEmissions.Add(estimationResource.CarbonEmissions)
		estimationTotal.ResourcesCount += 1
	}

	return EstimationReport{
		Info: EstimationInfo{
			UnitTime:                viper.Get("unit.time").(string),
			UnitWattTime:            fmt.Sprintf("%s%s", viper.Get("unit.power"), viper.Get("unit.time")),
			UnitCarbonEmissionsTime: fmt.Sprintf("%sCO2eq/%s", viper.Get("unit.carbon"), viper.Get("unit.time")),
			DateTime:                time.Now(),
		},
		Resources: estimationResources,
		Total:     estimationTotal,
	}

}

func EstimateResource(resource resources.ComputeResource) (*EstimationResource, *providers.UnsupportedProviderError) {
	switch resource.Provider {
	case providers.GCP:
		return estimateGCP(resource), nil
	}
	return nil, &providers.UnsupportedProviderError{Provider: resource.Provider.String()}
}

// Get the carbon emissions of a GCP resource
func estimateGCP(resource resources.ComputeResource) *EstimationResource {
	// Electric power used per unit of time
	avgWatt := EstimateWattHourGCP(resource) // Watt hour
	if viper.Get("unit.power").(string) == "kW" {
		avgWatt = avgWatt.Div(decimal.NewFromInt(1000))
	}
	if viper.Get("unit.time").(string) == "m" {
		avgWatt = avgWatt.Mul(decimal.NewFromInt(24 * 30))
	}
	if viper.Get("unit.time").(string) == "y" {
		avgWatt = avgWatt.Mul(decimal.NewFromInt(24 * 365))
	}

	// Regional grid emission per unit of time
	regionEmissions := GCPRegionEmission(resource.Region) // gCO2eq /kWh
	if viper.Get("unit.power").(string) == "W" {
		regionEmissions.GridCarbonIntensity = regionEmissions.GridCarbonIntensity.Div(decimal.NewFromInt(1000))
	}
	if viper.Get("unit.time").(string) == "m" {
		regionEmissions.GridCarbonIntensity = regionEmissions.GridCarbonIntensity.Mul(decimal.NewFromInt(24 * 30))
	}
	if viper.Get("unit.time").(string) == "y" {
		regionEmissions.GridCarbonIntensity = regionEmissions.GridCarbonIntensity.Mul(decimal.NewFromInt(24 * 365))
	}
	if viper.Get("unit.carbon").(string) == "kg" {
		regionEmissions.GridCarbonIntensity = regionEmissions.GridCarbonIntensity.Div(decimal.NewFromInt(1000))
	}

	// Carbon Emissions
	carbonEmissionPerTime := avgWatt.Mul(regionEmissions.GridCarbonIntensity)

	log.Debugf(
		"estimating resource %v.%v (%v): %v %v%v * %v %vCO2/%v%v = %v %vCO2/%v%v",
		resource.ResourceType,
		resource.Name,
		regionEmissions.Region,
		avgWatt,
		viper.Get("unit.power").(string),
		viper.Get("unit.time").(string),
		regionEmissions.GridCarbonIntensity,
		viper.Get("unit.carbon").(string),
		viper.Get("unit.power").(string),
		viper.Get("unit.time").(string),
		carbonEmissionPerTime,
		viper.Get("unit.carbon").(string),
		viper.Get("unit.power").(string),
		viper.Get("unit.time").(string),
	)

	return &EstimationResource{
		Resource:        resource,
		Power:           avgWatt.RoundFloor(10),
		CarbonEmissions: carbonEmissionPerTime.RoundFloor(10),
		AverageCPUUsage: decimal.NewFromFloat(viper.GetFloat64("avg_cpu_use")).RoundFloor(10),
	}
}
