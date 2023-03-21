package tfrefs

import (
	"github.com/carboniferio/carbonifer/internal/resources"
	tfjson "github.com/hashicorp/terraform-json"
)

type References struct {
	ResourceConfigs    map[string]*tfjson.ConfigResource
	ResourceReferences map[string]*tfjson.StateResource
	DataResources      map[string]resources.DataResource
	ProviderConfigs    map[string]string
}
