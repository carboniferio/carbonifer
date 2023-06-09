package providers

import (
	"strings"

	"github.com/carboniferio/carbonifer/internal/data"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/yunabe/easycsv"
)

var wattPerGPU map[string]GPUWatt

type GPUWatt struct {
	Name     string
	MinWatts decimal.Decimal
	MaxWatts decimal.Decimal
}

type gpuWattCSV struct {
	Name     string  `name:"name"`
	MinWatts float64 `name:"min watts"`
	MaxWatts float64 `name:"max watts"`
}

// Source: https://www.cloudcarbonfootprint.org/docs/methodology#appendix-iii-gpus-and-minmax-watts
func GetGPUWatt(gpuName string) GPUWatt {
	log.Debugf("  Getting info for GPU type: %v", gpuName)
	if wattPerGPU == nil {
		// Read the CSV records
		var records []gpuWattCSV
		gpuPowerDataFile := data.ReadDataFile("gpu_watt.csv")
		log.Debugf("  reading gpu power data from: %v", gpuPowerDataFile)
		if err := easycsv.NewReader(strings.NewReader(string(gpuPowerDataFile))).ReadAll(&records); err != nil {
			log.Fatal(err)
		}

		// Create a map to store the data
		wattPerGPU = make(map[string]GPUWatt)

		// Iterate over the records and add them to the map
		for _, record := range records {
			wattPerGPU[strings.ToLower(record.Name)] = GPUWatt{
				Name:     record.Name,
				MinWatts: decimal.NewFromFloat(record.MinWatts),
				MaxWatts: decimal.NewFromFloat(record.MaxWatts),
			}
		}
	}
	return wattPerGPU[strings.ToLower(gpuName)]
}
