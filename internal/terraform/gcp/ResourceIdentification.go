package gcp

import (
	"fmt"
	"strings"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	tfjson "github.com/hashicorp/terraform-json"
)

func getResourceIdentification(resource tfjson.StateResource) *resources.ResourceIdentification {
	region := resource.AttributeValues["region"]
	if region == nil {
		zone := resource.AttributeValues["zone"]
		replica_zones := resource.AttributeValues["replica_zones"]
		if zone != nil {
			region = strings.Join(strings.Split(zone.(string), "-")[:2], "-")
		} else if replica_zones != nil {
			region = strings.Join(strings.Split(replica_zones.([]interface{})[0].(string), "-")[:2], "-")
		} else {
			region = ""
		}
	}

	selfLink := ""
	if resource.AttributeValues["self_link"] != nil {
		selfLink = resource.AttributeValues["self_link"].(string)
	}

	name := resource.Name
	if resource.Index != nil {
		name = fmt.Sprintf("%v[%v]", resource.Name, resource.Index)
	}

	return &resources.ResourceIdentification{
		Name:         name,
		ResourceType: resource.Type,
		Provider:     providers.GCP,
		Region:       fmt.Sprint(region),
		SelfLink:     selfLink,
		Count:        1,
	}
}
