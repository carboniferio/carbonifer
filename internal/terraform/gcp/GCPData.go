package gcp

import (
	"fmt"
	"strings"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	tfjson "github.com/hashicorp/terraform-json"
)

func GetDataResource(tfResource tfjson.StateResource) resources.DataResource {
	resourceId := getDataResourceIdentification(tfResource)
	if resourceId.ResourceType == "google_compute_image" {
		diskSize := tfResource.AttributeValues["disk_size_gb"]
		diskSizeGb, ok := diskSize.(float64)
		specs := resources.DataImageSpecs{
			DiskSizeGb: diskSizeGb,
		}
		if ok {
			return resources.DataImageResource{
				Identification: resourceId,
				DataImageSpecs: []*resources.DataImageSpecs{&specs},
			}
		}
	}
	return nil
}

func getDataResourceIdentification(resource tfjson.StateResource) *resources.ResourceIdentification {
	region := resource.AttributeValues["region"]
	if region == nil {
		if resource.AttributeValues["zone"] != nil {
			zone := resource.AttributeValues["zone"].(string)
			region = strings.Join(strings.Split(zone, "-")[:2], "-")
		} else if resource.AttributeValues["replica_zones"] != nil {
			replica_zones := resource.AttributeValues["replica_zones"].([]interface{})
			// should be all in the same region
			region = strings.Join(strings.Split(replica_zones[0].(string), "-")[:2], "-")
		} else {
			region = ""
		}
	}

	return &resources.ResourceIdentification{
		Name:         resource.Name,
		ResourceType: resource.Type,
		Provider:     providers.GCP,
		Region:       fmt.Sprint(region),
	}
}
