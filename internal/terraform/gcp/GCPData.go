package gcp

import (
	"fmt"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/carboniferio/carbonifer/internal/utils"
	"github.com/tidwall/gjson"
)

func GetDataResource(tfResource *gjson.Result) resources.DataResource {
	resourceId := getDataResourceIdentification(tfResource)
	if resourceId.ResourceType == "google_compute_image" {
		fmt.Println(tfResource)
		diskSizeGbRes := utils.GetOr(tfResource, []string{"values.disk_size_gb", "values.size"})
		diskSizeGb := diskSizeGbRes.Float()
		specs := resources.DataImageSpecs{
			DiskSizeGb: diskSizeGb,
		}
		return resources.DataImageResource{
			Identification: resourceId,
			DataImageSpecs: []*resources.DataImageSpecs{&specs},
		}

	}
	return nil
}

func getDataResourceIdentification(resource *gjson.Result) *resources.ResourceIdentification {

	region := GetRegion(resource)

	return &resources.ResourceIdentification{
		Name:         resource.Get("name").String(),
		ResourceType: resource.Get("type").String(),
		Provider:     providers.GCP,
		Region:       fmt.Sprint(region),
	}
}
