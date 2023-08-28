package coefficients

import (
	"errors"
	"fmt"
	"strings"

	"github.com/carboniferio/carbonifer/internal/data"
	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"github.com/yunabe/easycsv"
)

// EmissionsPerRegion is a map of regions to their emissions
var EmissionsPerRegion map[string]Emissions

// Emissions is the emissions of a region
type Emissions struct {
	Region              string
	Location            string
	GridCarbonIntensity decimal.Decimal
}

// RegionEmission returns the emissions of a region
func RegionEmission(provider providers.Provider, region string) (*Emissions, error) {
	var dataFile string
	switch provider {
	case providers.AWS:
		dataFile = "aws_co2_region.csv"
	case providers.GCP:
		dataFile = "gcp_co2_region.csv"
	default:
		return nil, errors.New("Provider not supported")
	}
	if EmissionsPerRegion == nil {
		EmissionsPerRegion = loadEmissionsPerRegion(dataFile)
	}
	if region == "" {
		return nil, errors.New("Region cannot be empty")
	}
	emissions, ok := EmissionsPerRegion[region]
	if !ok {
		return nil, errors.New(fmt.Sprint("Region does not exist: ", region))
	}
	return &emissions, nil
}

type emissionsCSV struct {
	Region              string  `name:"Region"`
	Location            string  `name:"Location"`
	GridCarbonIntensity float64 `name:"Grid carbon intensity (gCO2eq / kWh)"`
}

// Source: Google
func loadEmissionsPerRegion(dataFile string) map[string]Emissions {
	// Read the CSV records
	var records []emissionsCSV
	regionEmissionFile := data.ReadDataFile(dataFile)
	log.Debugf("reading GCP region/grid emissions from: %v", dataFile)
	if err := easycsv.NewReader(strings.NewReader(string(regionEmissionFile))).ReadAll(&records); err != nil {
		log.Fatal(err)
	}

	// Create a map to store the data
	data := make(map[string]Emissions)

	// Iterate over the records and add them to the map
	for _, record := range records {

		data[record.Region] = Emissions{
			Region:              record.Region,
			Location:            record.Location,
			GridCarbonIntensity: decimal.NewFromFloat(record.GridCarbonIntensity),
		}
	}
	return data
}
