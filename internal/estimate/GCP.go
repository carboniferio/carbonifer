package estimate

import (
	"errors"
	"fmt"
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

func EstimateWattHourGCP(resource *resources.ComputeResource) decimal.Decimal {
	cpuEstimationInWh := estimateWattGCPCPU(resource)
	log.Debugf("%v.%v CPU in Wh: %v", resource.Identification.ResourceType, resource.Identification.Name, cpuEstimationInWh)
	memoryEstimationInWH := estimateWattMem(resource)
	log.Debugf("%v.%v Memory in Wh: %v", resource.Identification.ResourceType, resource.Identification.Name, memoryEstimationInWH)
	storageInWh := estimateWattStorage(resource)
	log.Debugf("%v.%v Storage in Wh: %v", resource.Identification.ResourceType, resource.Identification.Name, storageInWh)
	pue := GetEnergyCoefficients().GCP.PueAverage
	log.Debugf("%v.%v PUE %v", resource.Identification.ResourceType, resource.Identification.Name, pue)
	rawCarbonEstimate := decimal.Sum(
		cpuEstimationInWh,
		memoryEstimationInWH,
		storageInWh,
	)
	replicationFactor := resource.Specs.ReplicationFactor
	if replicationFactor == 0 {
		replicationFactor = 1
	}
	carbonEstimateIngCO2h := pue.Mul(rawCarbonEstimate).Mul(decimal.NewFromInt32(replicationFactor))
	log.Debugf("%v.%v Carbon in gCO2/h: %v", resource.Identification.ResourceType, resource.Identification.Name, carbonEstimateIngCO2h)
	return carbonEstimateIngCO2h
}

func estimateWattMem(resource *resources.ComputeResource) decimal.Decimal {
	return decimal.NewFromInt32(resource.Specs.MemoryMb).Div(decimal.NewFromInt32(1024)).Mul(GetEnergyCoefficients().GCP.MemoryWhGb)
}

func estimateWattGCPCPU(resource *resources.ComputeResource) decimal.Decimal {
	// Get average CPU usage
	averageCPUUse := decimal.NewFromFloat(viper.GetFloat64("provider.gcp.avg_cpu_use"))

	var avgWatts decimal.Decimal
	// Average Watts = Min Watts + Avg vCPU Utilization * (Max Watts - Min Watts)
	cpu_platform := resource.Specs.CPUType
	if cpu_platform != "" {
		cpu_platform := providers.GetCPUWatt(strings.ToLower(cpu_platform))
		avgWatts = cpu_platform.MinWatts.Add(averageCPUUse.Mul(cpu_platform.MaxWatts.Sub(cpu_platform.MinWatts)))
	} else {
		avgWatts = GetEnergyCoefficients().GCP.CPUMinWh.Add(averageCPUUse.Mul(GetEnergyCoefficients().GCP.CPUMaxWh.Sub(GetEnergyCoefficients().GCP.CPUMinWh)))
	}
	return avgWatts.Mul(decimal.NewFromInt32(resource.Specs.VCPUs))
}

func estimateWattStorage(resource *resources.ComputeResource) decimal.Decimal {
	provider := resource.Identification.Provider.String()
	storageSsdWhGb := GetEnergyCoefficients().GetByName(provider).StorageSsdWhTb.Div(decimal.NewFromInt32(1024))
	storageHddWhGb := GetEnergyCoefficients().GetByName(provider).StorageHddWhTb.Div(decimal.NewFromInt32(1024))
	storageSSDWh := resource.Specs.SsdStorage.Mul(storageSsdWhGb)
	storageHddWh := resource.Specs.HddStorage.Mul(storageHddWhGb)
	return storageSSDWh.Add(storageHddWh)
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

func GCPRegionEmission(region string) (*GCPEmissions, error) {
	if gcpEmissionsPerRegion == nil {
		gcpEmissionsPerRegion = loadEmissionsPerRegion()
	}
	if region == "" {
		return nil, errors.New("Region cannot be empty")
	}
	emissions, ok := gcpEmissionsPerRegion[region]
	if !ok {
		return nil, errors.New(fmt.Sprint("Region does not exist: ", region))
	}
	return &emissions, nil
}
