package terraform

import (
	"github.com/carboniferio/carbonifer/internal/resources"
	tfjson "github.com/hashicorp/terraform-json"
)

func GetDataResource(tfResource tfjson.StateResource) resources.DataResource {
	resourceId := getResourceIdentification(tfResource)
	if resourceId.ResourceType == "google_compute_image" {
		diskSize := tfResource.AttributeValues["disk_size_gb"]
		diskSizeGb, ok := diskSize.(float64)
		if ok {
			return resources.DataImageResource{
				Identification: resourceId,
				DataImageSpecs: &resources.DataImageSpecs{
					DiskSizeGb: diskSizeGb,
				},
			}
		}
	}
	return nil
}
