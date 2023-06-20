package gcp

import (
	"fmt"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/tidwall/gjson"
)

func getResourceIdentification(resource *gjson.Result) *resources.ResourceIdentification {
	region := GetRegion(resource)

	name := resource.Get("name").String()
	if resource.Get("index").Exists() {
		name = fmt.Sprintf("%v[%v]", resource.Get("name").String(), resource.Get("index").Int())
	}

	return &resources.ResourceIdentification{
		Name:         name,
		ResourceType: resource.Get("type").String(),
		Provider:     providers.GCP,
		Region:       fmt.Sprint(region),
		Count:        1,
	}
}
