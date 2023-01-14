package estimate

import (
	"path/filepath"
	"strings"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/spf13/viper"
	"github.com/yunabe/easycsv"
)

// Source: https://www.cloudcarbonfootprint.org/docs/methodology/#appendix-i-energy-coefficients
// in Watt Hour

var gcpEmissionsPerRegion map[string]GCPEmissions

func EstimateWattHourGCP(resource resources.ComputeResource) decimal.Decimal {
	cpuEstimationInWh := estimateWattGCPCPU(resource)
	log.Debugf("%v.%v CPU in Wh: %v", resource.ResourceType, resource.Name, cpuEstimationInWh)
	memoryEstimationInWH := estimateWattMem(resource)
	log.Debugf("%v.%v Memory in Wh: %v", resource.ResourceType, resource.Name, memoryEstimationInWH)
	pue := GetEnergyCoefficients().GCP.PueAverage
	log.Debugf("%v.%v PUE %v", resource.ResourceType, resource.Name, pue)
	rawCarbonEstimate := cpuEstimationInWh.Add(memoryEstimationInWH)
	carbonEstimateIngCO2h := pue.Mul(rawCarbonEstimate)
	log.Debugf("%v.%v Carbon in gCO2/h: %v", resource.ResourceType, resource.Name, carbonEstimateIngCO2h)
	return carbonEstimateIngCO2h
}

func estimateWattMem(resource resources.ComputeResource) decimal.Decimal {
	return decimal.NewFromInt32(resource.MemoryMb).Div(decimal.NewFromInt32(1024)).Mul(GetEnergyCoefficients().GCP.MemoryWhGb)
}

func estimateWattGCPCPU(resource resources.ComputeResource) decimal.Decimal {
	// Get average CPU usage
	averageCPUUse := decimal.NewFromFloat(viper.GetFloat64("avg_cpu_use"))

	var avgWatts decimal.Decimal
	// Average Watts = Min Watts + Avg vCPU Utilization * (Max Watts - Min Watts)
	cpu_platform := resource.CPUType
	if cpu_platform != "" {
		cpu_platform := providers.GetCPUWatt(strings.ToLower(cpu_platform))
		avgWatts = cpu_platform.MinWatts.Add(averageCPUUse.Mul(cpu_platform.MaxWatts.Sub(cpu_platform.MinWatts)))
	} else {
		avgWatts = GetEnergyCoefficients().GCP.CPUMinWh.Add(averageCPUUse.Mul(GetEnergyCoefficients().GCP.CPUMaxWh.Sub(GetEnergyCoefficients().GCP.CPUMinWh)))
	}
	return avgWatts.Mul(decimal.NewFromInt32(resource.VCPUs))
}

type gcpEmissionsCSV struct {
	Region              string  `name:"Google Cloud Region"`
	Location            string  `name:"Location"`
	GoogleCFE           string  `name:"Google CFE"`
	GridCarbonIntensity float64 `name:"Grid carbon intensity (gCO2eq / kWh)"`
	NetCarbonEmissions  float64 `name:"Google Cloud net carbon emissions"`
}

type GCPEmissions struct {
	Region              string
	Location            string
	GoogleCFE           string
	GridCarbonIntensity decimal.Decimal
	NetCarbonEmissions  decimal.Decimal
}

// Source: Google
func loadEmissionsPerRegion() map[string]GCPEmissions {
	// Read the CSV records
	var records []gcpEmissionsCSV
	gcpRegionEmissionFile := filepath.Join(viper.GetString("data.path"), "gcp_watt_region.csv")
	log.Debugf("reading GCP region/grid emissions from: %v", gcpRegionEmissionFile)
	if err := easycsv.NewReaderFile(gcpRegionEmissionFile).ReadAll(&records); err != nil {
		log.Fatal(err)
	}

	// Create a map to store the data
	data := make(map[string]GCPEmissions)

	// Iterate over the records and add them to the map
	for _, record := range records {

		data[record.Region] = GCPEmissions{
			Region:              record.Region,
			Location:            record.Location,
			GoogleCFE:           record.GoogleCFE,
			GridCarbonIntensity: decimal.NewFromFloat(record.GridCarbonIntensity),
			NetCarbonEmissions:  decimal.NewFromFloat(record.NetCarbonEmissions),
		}
	}
	return data
}

func GCPRegionEmission(region string) GCPEmissions {
	if gcpEmissionsPerRegion == nil {
		gcpEmissionsPerRegion = loadEmissionsPerRegion()
	}
	return gcpEmissionsPerRegion[region]
}
