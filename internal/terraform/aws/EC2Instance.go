package aws

import (
	"github.com/carboniferio/carbonifer/internal/providers/aws"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/carboniferio/carbonifer/internal/terraform/tfrefs"
	tfjson "github.com/hashicorp/terraform-json"
)

func getEC2Instance(
	resource tfjson.StateResource,
	tfRefs *tfrefs.References, groupZone interface{}) *resources.ComputeResourceSpecs {

	instanceType := resource.AttributeValues["instance_type"].(string)

	machineType := aws.GetAWSInstanceType(instanceType)

	return &resources.ComputeResourceSpecs{
		VCPUs:             machineType.VCPU,
		MemoryMb:          machineType.MemoryMb,
		ReplicationFactor: 1,
	}
}
