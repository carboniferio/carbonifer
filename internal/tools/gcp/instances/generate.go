package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"google.golang.org/api/compute/v1"

	"github.com/carboniferio/carbonifer/internal/providers/gcp"
	toolsgcp "github.com/carboniferio/carbonifer/internal/tools/gcp"
)

var cpuTypes map[string]machineFamily

type machineFamily struct {
	Name         string   `json:"Name"`
	CPUTypes     []string `json:"CPU types"`
	Architecture string   `json:"Architecture"`
}

func getCPUTypes(machineType string) []string {
	if cpuTypes == nil {
		// cpu_types.json manually generated from: https://cloud.google.com/compute/docs/machine-resource
		jsonFile, err := os.Open("internal/tools/gcp/instances/cpu_types.json")
		if err != nil {
			log.Fatal(err)
		}

		cpuTypes = make(map[string]machineFamily)
		byteValue, _ := io.ReadAll(jsonFile)
		err = json.Unmarshal(byteValue, &cpuTypes)
		if err != nil {
			log.Panic(err)
		}

		// defer the closing of our jsonFile so that we can parse it later on
		defer jsonFile.Close()
	}
	family := strings.Split(machineType, "-")[0]

	familyTypes, ok := cpuTypes[family]
	if !ok {
		return nil
	}

	return familyTypes.CPUTypes
}

func getMachineTypesForZone(client *compute.Service, project string, zone string) map[string]gcp.MachineType {
	machineTypes := make(map[string]gcp.MachineType)
	// Get the list of available machine types
	machineTypesArray, err := client.MachineTypes.List(project, zone).Do()
	if err != nil {
		log.Fatalf("Error getting machine type list: %v", err)
	}

	for _, machineType := range machineTypesArray.Items {
		_, ok := machineTypes[machineType.Name]
		if ok {
			log.Fatalf("There is already a machine type %v", machineType.Name)
		}
		machineTypes[machineType.Name] = gcp.MachineType{
			Name:     machineType.Name,
			Vcpus:    int32(machineType.GuestCpus),
			MemoryMb: int32(machineType.MemoryMb),
			CPUTypes: getCPUTypes(machineType.Name),
			GPUTypes: getGPUs(machineType),
		}
	}
	return machineTypes
}

func getGPUs(machineType *compute.MachineType) []string {
	var gpuTypes []string
	for _, accelerator := range machineType.Accelerators {
		for i := 0; i < int(accelerator.GuestAcceleratorCount); i++ {
			gpuTypes = append(gpuTypes, accelerator.GuestAcceleratorType)
		}
	}
	return gpuTypes
}

func main() {
	if len(os.Args) < 2 {
		generateGlobal()
		return
	}

	command := os.Args[1]

	switch command {
	case "global":
		generateGlobal()
	case "regions":
		generatePerRegion()
	default:
		generateGlobal()
	}
}

func generatePerRegion() {
	machineTypesByZone := retrieveData()
	// Generate the JSON representation of the list
	jsonData, err := json.MarshalIndent(machineTypesByZone, "", "  ")
	if err != nil {
		log.Fatalf("Error generating JSON: %v", err)
	}
	fmt.Println(string(jsonData))
}

func generateGlobal() {
	machineTypesByZone := retrieveData()
	machineTypes := map[string]gcp.MachineType{}
	for _, machineTypesList := range *machineTypesByZone {
		for name, machineType := range machineTypesList {
			machineTypes[name] = machineType
		}
	}
	// Generate the JSON representation of the list
	jsonData, err := json.MarshalIndent(machineTypes, "", "  ")
	if err != nil {
		log.Fatalf("Error generating JSON: %v", err)
	}
	fmt.Println(string(jsonData))
}

func retrieveData() *map[string]map[string]gcp.MachineType {

	ctx := context.Background()

	// Create a new Compute Engine client
	client, err := compute.NewService(ctx)
	if err != nil {
		log.Fatalf("Error creating Compute Engine client: %v", err)
	}

	project := toolsgcp.GetProjectID()

	machineTypesByZone := make(map[string]map[string]gcp.MachineType)
	zones, err := client.Zones.List(project).Do()
	if err != nil {
		log.Fatal(err)
	}
	for _, zone := range zones.Items {
		machineTypesByZone[zone.Name] = getMachineTypesForZone(client, project, zone.Name)
	}
	return &machineTypesByZone
}
