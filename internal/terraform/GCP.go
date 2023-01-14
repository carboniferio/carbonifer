package terraform

import (
	"strings"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	tfjson "github.com/hashicorp/terraform-json"
)

func GetResource(tfResource tfjson.StateResource) resources.Resource {
	resourceId := getResourceIdentification(tfResource)
	if resourceId.ResourceType == "google_compute_instance" {
		specs := getComputeResourceSpecs(tfResource)
		return resources.ComputeResource{
			Identification: resourceId,
			Specs:          specs,
		}
	}
	return resources.UnsupportedResource{
		Identification: resourceId,
	}
}

func getResourceIdentification(resource tfjson.StateResource) *resources.ComputeResourceIdentification {
	var region string
	if resource.AttributeValues["zone"] != nil {
		zone := resource.AttributeValues["zone"].(string)
		region = strings.Join(strings.Split(zone, "-")[:2], "-")
	}

	return &resources.ComputeResourceIdentification{
		Name:         resource.Name,
		ResourceType: resource.Type,
		Provider:     providers.GCP,
		Region:       region,
	}
}

func getComputeResourceSpecs(resource tfjson.StateResource) *resources.ComputeResourceSpecs {
	machine_type := resource.AttributeValues["machine_type"].(string)
	zone := resource.AttributeValues["zone"].(string)
	machineType := providers.GetGCPMachineType(machine_type, zone)
	CPUType, ok := resource.AttributeValues["cpu_platform"].(string)
	if !ok {
		CPUType = ""
	}
	return &resources.ComputeResourceSpecs{
		Gpu:      machineType.Gpus,
		VCPUs:    machineType.Vcpus,
		MemoryMb: machineType.MemoryMb,
		CPUType:  CPUType,
	}
}
