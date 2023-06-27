package aws

import (
	"fmt"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/carboniferio/carbonifer/internal/terraform/tfrefs"
	tfjson "github.com/hashicorp/terraform-json"
)

func getResourceIdentification(resource tfjson.StateResource, tfRefs *tfrefs.References) *resources.ResourceIdentification {
	region := resource.AttributeValues["region"]
	if region == nil {
		region = tfRefs.ProviderConfigs["region"]
	}

	name := resource.Name
	if resource.Index != nil {
		name = fmt.Sprintf("%v[%v]", resource.Name, resource.Index)
	}

	provider, _ := providers.ParseProvider(resource.ProviderName)

	return &resources.ResourceIdentification{
		Name:         name,
		ResourceType: resource.Type,
		Provider:     provider,
		Region:       fmt.Sprint(region),
		Count:        1,
	}
}
