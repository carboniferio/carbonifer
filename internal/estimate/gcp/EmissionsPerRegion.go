package gcp

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
	"github.com/yunabe/easycsv"
)

var gcpEmissionsPerRegion map[string]GCPEmissions

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
