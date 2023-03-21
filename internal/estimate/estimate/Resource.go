package estimate

import (
	"github.com/carboniferio/carbonifer/internal/estimate/coefficients"
	"github.com/carboniferio/carbonifer/internal/estimate/estimation"

	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Get the carbon emissions of a GCP resource
func EstimateSupportedResource(resource resources.Resource) *estimation.EstimationResource {

	var computeResource resources.ComputeResource = resource.(resources.ComputeResource)
	// Electric power used per unit of time
	avgWatt := estimateWattHour(&computeResource) // Watt hour
	if viper.Get("unit.power").(string) == "kW" {
		avgWatt = avgWatt.Div(decimal.NewFromInt(1000))
	}
	if viper.Get("unit.time").(string) == "m" {
		avgWatt = avgWatt.Mul(decimal.NewFromInt(24 * 30))
	}
	if viper.Get("unit.time").(string) == "y" {
		avgWatt = avgWatt.Mul(decimal.NewFromInt(24 * 365))
	}
	avgWattStr := avgWatt.String()

	// Regional grid emission per unit of time
	regionEmissions, err := coefficients.RegionEmission(resource.GetIdentification().Provider, resource.GetIdentification().Region) // gCO2eq /kWh
	if err != nil {
		log.Fatalf("Error while getting region emissions for %v: %v", resource.GetAddress(), err)
	}
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
	carbonEmissionPerTimeStr := carbonEmissionPerTime.String()

	log.Debugf(
		"estimating resource %v.%v (%v): %v %v%v * %v %vCO2/%v%v = %v %vCO2/%v%v * %v",
		computeResource.Identification.ResourceType,
		computeResource.Identification.Name,
		regionEmissions.Region,
		avgWattStr,
		viper.Get("unit.power").(string),
		viper.Get("unit.time").(string),
		regionEmissions.GridCarbonIntensity,
		viper.Get("unit.carbon").(string),
		viper.Get("unit.power").(string),
		viper.Get("unit.time").(string),
		carbonEmissionPerTimeStr,
		viper.Get("unit.carbon").(string),
		viper.Get("unit.power").(string),
		viper.Get("unit.time").(string),
		resource.GetIdentification().Count,
	)

	est := &estimation.EstimationResource{
		Resource:        &computeResource,
		Power:           avgWatt.RoundFloor(10),
		CarbonEmissions: carbonEmissionPerTime.RoundFloor(10),
		AverageCPUUsage: decimal.NewFromFloat(viper.GetFloat64("provider.gcp.avg_cpu_use")).RoundFloor(10),
		Count:           decimal.NewFromInt(int64(computeResource.Identification.Count)),
	}
	return est
}
