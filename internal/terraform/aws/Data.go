package aws

import (
	"strconv"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/carboniferio/carbonifer/internal/terraform/tfrefs"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/tidwall/gjson"
)

func GetDataResource(tfResource *gjson.Result) resources.DataResource {
	resourceId := getDataResourceIdentification(tfResource)
	if resourceId.ResourceType == "aws_ami" {
		diskMappingI := tfResource.AttributeValues["block_device_mappings"]
		if diskMappingI != nil {
			diskMapping := diskMappingI.([]interface{})
			specs := make([]*resources.DataImageSpecs, len(diskMapping))
			for i, disk := range diskMapping {
				ebs := disk.(map[string]interface{})["ebs"].(map[string]interface{})
				if ebs["volume_size"] != nil {
					diskSizeGb, _ := strconv.ParseFloat(ebs["volume_size"].(string), 64)
					volumeType := ""
					if ebs["volume_type"] != nil {
						volumeType = ebs["volume_type"].(string)
					}
					diskSpecs := resources.DataImageSpecs{
						DiskSizeGb: diskSizeGb,
						DeviceName: disk.(map[string]interface{})["device_name"].(string),
						VolumeType: volumeType,
					}
					specs[i] = &diskSpecs
				}
			}
			return resources.EbsDataResource{
				Identification: resourceId,
				DataImageSpecs: specs,
				AwsId:          tfResource.AttributeValues["id"].(string),
			}
		}
	}
	if resourceId.ResourceType == "aws_ebs_snapshot" {
		diskSize := tfResource.AttributeValues["volume_size"]
		diskSizeGb := diskSize.(float64)
		return resources.EbsDataResource{
			Identification: resourceId,
			DataImageSpecs: []*resources.DataImageSpecs{
				{
					DiskSizeGb: diskSizeGb,
				},
			},
			AwsId: tfResource.AttributeValues["id"].(string),
		}
	}

	return resources.DataImageResource{
		Identification: resourceId,
	}
}

func getDataResourceIdentification(resource tfjson.StateResource) *resources.ResourceIdentification {

	return &resources.ResourceIdentification{
		Name:         resource.Name,
		ResourceType: resource.Type,
		Provider:     providers.AWS,
	}
}

func getAwsImage(tfRefs *tfrefs.References, awsImageId string) *resources.EbsDataResource {
	imageI := tfRefs.DataResources[awsImageId]

	var image *resources.EbsDataResource
	if imageI != nil {
		i := imageI.(resources.EbsDataResource)
		image = &i
	}
	return image
}
