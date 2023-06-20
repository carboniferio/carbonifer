package aws

import (
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/carboniferio/carbonifer/internal/terraform/tfrefs"
	"github.com/tidwall/gjson"
)

func GetResource(
	tfResource *gjson.Result,
	tfRefs *tfrefs.References) resources.Resource {

	resourceId := getResourceIdentification(tfResource, tfRefs)
	if resourceId.ResourceType == "aws_instance" {
		specs := getEC2Instance(tfResource, tfRefs)
		return resources.ComputeResource{
			Identification: resourceId,
			Specs:          specs,
		}
	}
	if resourceId.ResourceType == "aws_ebs_volume" {
		specs := getEbsVolume(tfResource, tfRefs)
		return resources.ComputeResource{
			Identification: resourceId,
			Specs:          specs,
		}
	}

	return resources.UnsupportedResource{
		Identification: resourceId,
	}
}
