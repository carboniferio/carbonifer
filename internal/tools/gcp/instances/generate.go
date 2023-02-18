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
	tools_gcp "github.com/carboniferio/carbonifer/internal/tools/gcp"
)

const DEFAULT_ZONE = "us-central1-a"

var cpuTypes map[string]MachineFamily

type MachineFamily struct {
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

		cpuTypes = make(map[string]MachineFamily)
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
			CpuTypes: getCPUTypes(machineType.Name),
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

	ctx := context.Background()

	// Create a new Compute Engine client
	client, err := compute.NewService(ctx)
	if err != nil {
		log.Fatalf("Error creating Compute Engine client: %v", err)
	}

	project := tools_gcp.GetProjectId()

	machineTypesByZone := make(map[string]map[string]gcp.MachineType)
	zones, err := client.Zones.List(project).Do()
	if err != nil {
		log.Fatal(err)
	}
	for _, zone := range zones.Items {
		machineTypesByZone[zone.Name] = getMachineTypesForZone(client, project, zone.Name)
	}

	// Generate the JSON representation of the list
	jsonData, err := json.MarshalIndent(machineTypesByZone, "", "  ")
	if err != nil {
		log.Fatalf("Error generating JSON: %v", err)
	}

	fmt.Println(string(jsonData))
}
