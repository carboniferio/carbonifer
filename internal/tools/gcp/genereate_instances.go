package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"

	"github.com/carboniferio/carbonifer/internal/providers"
)

const DEFAULT_ZONE = "us-central1-a"

var cpuTypes map[string]MachineFamily

func getProjectId() string {
	ctx := context.Background()

	// Get the default client using the default credentials
	client, err := google.DefaultClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	crmService, err := cloudresourcemanager.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatal(err)
	}

	// Set the project ID
	projects, err := crmService.Projects.List().Do()
	if err != nil {
		log.Fatal(err)
	}
	return projects.Projects[0].Name
}

type MachineFamily struct {
	Name         string   `json:"Name"`
	CPUTypes     []string `json:"CPU types"`
	Architecture string   `json:"Architecture"`
}

func getCPUTypes(machineType string) []string {
	if cpuTypes == nil {
		// cpu_types.json manually generated from: https://cloud.google.com/compute/docs/machine-resource
		jsonFile, err := os.Open("cpu_types.json")
		if err != nil {
			log.Fatal(err)
		}

		cpuTypes = make(map[string]MachineFamily)
		byteValue, _ := ioutil.ReadAll(jsonFile)
		json.Unmarshal(byteValue, &cpuTypes)

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

func getMachineTypesForZone(client *compute.Service, project string, zone string) map[string]providers.MachineType {
	machineTypes := make(map[string]providers.MachineType)
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
		machineTypes[machineType.Name] = providers.MachineType{
			Name:     machineType.Name,
			Vcpus:    int32(machineType.GuestCpus),
			MemoryMb: int32(machineType.MemoryMb),
			CpuTypes: getCPUTypes(machineType.Name),
			Gpus:     getGPUs(machineType),
		}
	}
	return machineTypes
}

func getGPUs(machineType *compute.MachineType) int32 {
	var count int32 = 0
	for _, accelerator := range machineType.Accelerators {
		count += int32(accelerator.GuestAcceleratorCount)
	}
	return count
}

func main() {

	ctx := context.Background()

	// Create a new Compute Engine client
	client, err := compute.NewService(ctx)
	if err != nil {
		log.Fatalf("Error creating Compute Engine client: %v", err)
	}

	project := getProjectId()

	machineTypesByZone := make(map[string]map[string]providers.MachineType)
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
