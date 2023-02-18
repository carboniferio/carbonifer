package gcp

import (
	"fmt"
	"strings"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	tfjson "github.com/hashicorp/terraform-json"
)

func getResourceIdentification(resource tfjson.ConfigResource) *resources.ResourceIdentification {
	region := GetConstFromConfig(&resource, "region")
	if region == nil {
		zone := GetConstFromConfig(&resource, "zone")
		replica_zones := GetConstFromConfig(&resource, "replica_zones")
		if zone != nil {
			region = strings.Join(strings.Split(zone.(string), "-")[:2], "-")
		} else if replica_zones != nil {
			region = strings.Join(strings.Split(replica_zones.([]interface{})[0].(string), "-")[:2], "-")
		} else {
			region = ""
		}
	}
	selfLinkExpr := GetConstFromConfig(&resource, "self_link")
	var selfLink string
	if selfLinkExpr != nil {
		selfLink = GetConstFromConfig(&resource, "self_link").(string)
	}

	return &resources.ResourceIdentification{
		Name:         resource.Name,
		ResourceType: resource.Type,
		Provider:     providers.GCP,
		Region:       fmt.Sprint(region),
		SelfLink:     selfLink,
		Count:        1,
	}
}
