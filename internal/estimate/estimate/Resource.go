package estimate

import (
	"strings"

	"github.com/carboniferio/carbonifer/internal/estimate/coefficients"
	"github.com/carboniferio/carbonifer/internal/estimate/estimation"

	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// EstimateSupportedResource gets the carbon emissions of a GCP resource
func EstimateSupportedResource(resource resources.Resource) *estimation.EstimationResource {

	var computeResource resources.ComputeResource = resource.(resources.ComputeResource)
	// Electric power used per unit of time
	// It's computed first in watt per hour
	avgWattHour := estimateWattHour(&computeResource) // Watt hour
	avgKWattHour := avgWattHour.Div(decimal.NewFromInt(1000))

	// Regional grid emission per unit of time
	// regionEmissions are in gCO2/kWh
	regionEmissions, err := coefficients.RegionEmission(resource.GetIdentification().Provider, resource.GetIdentification().Region) // gCO2eq /kWh
	if err != nil {
		log.Fatalf("Error while getting region emissions for %v: %v", resource.GetAddress(), err)
	}

	// Carbon Emissions
	carbonEmissionInGCO2PerH := avgKWattHour.Mul(regionEmissions.GridCarbonIntensity)
	carbonEmissionPerTime := carbonEmissionInGCO2PerH
	if strings.ToLower(viper.GetString("unit.time")) == "d" {
		carbonEmissionPerTime = carbonEmissionPerTime.Mul(decimal.NewFromInt(24))
	}
	if strings.ToLower(viper.GetString("unit.time")) == "m" {
		carbonEmissionPerTime = carbonEmissionPerTime.Mul(decimal.NewFromInt(24 * 30))
	}
	if strings.ToLower(viper.GetString("unit.time")) == "y" {
		carbonEmissionPerTime = carbonEmissionPerTime.Mul(decimal.NewFromInt(24 * 365))
	}
	if strings.ToLower(viper.GetString("unit.carbon")) == "kg" {
		carbonEmissionPerTime = carbonEmissionPerTime.Div(decimal.NewFromInt(1000))
	}
	carbonEmissionPerTimeStr := carbonEmissionPerTime.String()

	log.Debugf(
		"estimating resource %v.%v (%v): %v %v%v * %v %vCO2/%v%v = %v %vCO2/%v%v * %v = %v %vCO2/%v%v * %v",
		computeResource.Identification.ResourceType,
		computeResource.Identification.Name,
		regionEmissions.Region,
		avgKWattHour.String(),
		"kW",
		"h",
		regionEmissions.GridCarbonIntensity,
		"g",
		"kW",
		"h",
		carbonEmissionInGCO2PerH,
		"g",
		"kW",
		"h",
		resource.GetIdentification().Count,
		carbonEmissionPerTimeStr,
		viper.GetString("unit.carbon"),
		viper.GetString("unit.power"),
		viper.GetString("unit.time"),
		resource.GetIdentification().Count,
	)

	if resource.GetIdentification().Name == "my_cluster_autoscaled" {
		log.Println("my_cluster_autoscaled")
	}

	count := int64(computeResource.Identification.Count)
	replicationFactor := int64(computeResource.Identification.ReplicationFactor)

	est := &estimation.EstimationResource{
		Resource:        &computeResource,
		Power:           avgWattHour.RoundFloor(10),
		CarbonEmissions: carbonEmissionPerTime.RoundFloor(10),
		AverageCPUUsage: decimal.NewFromFloat(viper.GetFloat64("provider.gcp.avg_cpu_use")).RoundFloor(10),
		TotalCount:      decimal.NewFromInt(count * replicationFactor),
	}
	return est
}
