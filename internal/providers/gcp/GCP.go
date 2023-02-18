package gcp

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/yunabe/easycsv"
)

type MachineType struct {
	Name     string   `json:"name"`
	Vcpus    int32    `json:"vcpus"`
	GPUTypes []string `json:"gpus"`
	MemoryMb int32    `json:"memoryMb"`
	CpuTypes []string `json:"cpuTypes"`
}

type SqlTier struct {
	Name        string `json:"name"`
	Vcpus       int64  `json:"vcpus"`
	MemoryMb    int64  `json:"memoryMb"`
	DiskQuotaGB int64  `json:"DiskQuotaGB"`
}

type CPUWatt struct {
	Architecture        string
	MinWatts            decimal.Decimal
	MaxWatts            decimal.Decimal
	GridCarbonIntensity decimal.Decimal
}

var gcpInstanceTypes map[string]map[string]MachineType
var gcpWattPerCPU map[string]CPUWatt
var gcpSQLTiers map[string]SqlTier

func GetGCPMachineType(machineTypeStr string, zone string) MachineType {
	log.Debugf("  Getting info for GCP machine type: %v", machineTypeStr)
	// Custom format is custom-<number_cpus>-<ram_mb>
	customMachineRegex := regexp.MustCompile(`custom-(?P<vcpus>\d+)-(?P<mem>\d+)(-ext)?`)
	if customMachineRegex.MatchString(machineTypeStr) {
		log.Debugf("  custom machine: %v", machineTypeStr)
		customValues := customMachineRegex.FindAllStringSubmatch(machineTypeStr, -1)[0]
		if len(customValues) < 3 {
			log.Fatalf("GCP Custom machine name malformed : %v", machineTypeStr)
		}
		vCPUs, err := strconv.Atoi(customValues[1])
		if err != nil {
			log.Fatalf(err.Error())
		}
		ram, err := strconv.Atoi(customValues[2])
		if err != nil {
			log.Fatalf(err.Error())
		}
		return MachineType{
			Name:     machineTypeStr,
			Vcpus:    int32(vCPUs),
			MemoryMb: int32(ram),
		}
	}
	if gcpInstanceTypes == nil {
		gcpInstancesDataFile := filepath.Join(viper.GetString("data.path"), "gcp_instances.json")
		log.Debugf("  reading gcp instances data from: %v", gcpInstancesDataFile)
		jsonFile, err := os.Open(gcpInstancesDataFile)
		if err != nil {
			log.Fatal(err)
		}
		defer jsonFile.Close()

		byteValue, _ := io.ReadAll(jsonFile)
		err = json.Unmarshal([]byte(byteValue), &gcpInstanceTypes)
		if err != nil {
			log.Fatal(err)
		}
	}

	return gcpInstanceTypes[zone][machineTypeStr]

}

type cpuWattCSV struct {
	Architecture        string  `name:"Architecture"`
	MinWatts            float64 `name:"Min Watts"`
	MaxWatts            float64 `name:"Max Watts"`
	GridCarbonIntensity float64 `name:"GB/Chip"`
}

// Source: https://github.com/cloud-carbon-footprint/cloud-carbon-coefficients/blob/5fcb96101c6f28dac5060f8794bca5d4da6c72d8/output/coefficients-gcp-use.csv
func GetCPUWatt(cpu string) CPUWatt {
	log.Debugf("  Getting info for GCP CPU type: %v", cpu)
	if gcpWattPerCPU == nil {
		// Read the CSV records
		var records []cpuWattCSV
		gcpPowerDataFile := filepath.Join(viper.GetString("data.path"), "gcp_watt_cpu.csv")
		log.Debugf("  reading GCP cpu power data from: %v", gcpPowerDataFile)
		if err := easycsv.NewReaderFile(gcpPowerDataFile).ReadAll(&records); err != nil {
			log.Fatal(err)
		}

		// Create a map to store the data
		gcpWattPerCPU = make(map[string]CPUWatt)

		// Iterate over the records and add them to the map
		for _, record := range records {
			gcpWattPerCPU[strings.ToLower(record.Architecture)] = CPUWatt{
				Architecture:        record.Architecture,
				MinWatts:            decimal.NewFromFloat(record.MinWatts),
				MaxWatts:            decimal.NewFromFloat(record.MaxWatts),
				GridCarbonIntensity: decimal.NewFromFloat(record.GridCarbonIntensity),
			}
		}
	}
	return gcpWattPerCPU[strings.ToLower(cpu)]
}

func GetGCPSQLTier(tierName string) SqlTier {
	log.Debugf("  Getting info for GCP SQL tier: %v", tierName)
	// Custom format db-custom-<number_cpus>-<ram_mb>
	customTierRegex := regexp.MustCompile(`db-custom-(?P<vcpus>\d+)-(?P<mem>\d+)`)
	if customTierRegex.MatchString(tierName) {
		log.Debugf("  custom SQL Tier: %v", tierName)
		customValues := customTierRegex.FindAllStringSubmatch(tierName, -1)[0]
		if len(customValues) < 3 {
			log.Fatalf("GCP Custom tier name malformed : %v", tierName)
		}
		vCPUs, err := strconv.Atoi(customValues[1])
		if err != nil {
			log.Fatalf(err.Error())
		}
		ram, err := strconv.Atoi(customValues[2])
		if err != nil {
			log.Fatalf(err.Error())
		}
		return SqlTier{
			Name:     tierName,
			Vcpus:    int64(vCPUs),
			MemoryMb: int64(ram),
		}
	}
	if gcpSQLTiers == nil {
		gcpSQLTierDataFile := filepath.Join(viper.GetString("data.path"), "gcp_sql_tiers.json")
		log.Debugf("  reading gcp sql tier data from: %v", gcpSQLTierDataFile)
		jsonFile, err := os.Open(gcpSQLTierDataFile)
		if err != nil {
			log.Fatal(err)
		}
		defer jsonFile.Close()

		byteValue, _ := io.ReadAll(jsonFile)
		err = json.Unmarshal([]byte(byteValue), &gcpSQLTiers)
		if err != nil {
			log.Fatal(err)
		}
	}

	return gcpSQLTiers[tierName]

}
