package terraform

import (
	"strings"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	tfjson "github.com/hashicorp/terraform-json"
)

func GetResource(resource tfjson.StateResource) *resources.ComputeResource {
	if resource.Type == "google_compute_instance" {
		return getComputeResource(resource)
	}
	return nil
}

func getComputeResource(resource tfjson.StateResource) *resources.ComputeResource {
	machine_type := resource.AttributeValues["machine_type"].(string)
	zone := resource.AttributeValues["zone"].(string)
	region := strings.Join(strings.Split(zone, "-")[:2], "-")
	machineType := providers.GetGCPMachineType(machine_type, zone)
	CPUType, ok := resource.AttributeValues["cpu_platform"].(string)
	if !ok {
		CPUType = ""
	}

	return &resources.ComputeResource{
		Name:         resource.Name,
		ResourceType: resource.Type,
		Provider:     providers.GCP,
		Region:       region,
		Gpu:          machineType.Gpus,
		VCPUs:        machineType.Vcpus,
		MemoryMb:     machineType.MemoryMb,
		CPUType:      CPUType,
	}

}
