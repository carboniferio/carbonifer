package tfrefs

import (
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/tidwall/gjson"
)

type References struct {
	ResourceConfigs    map[string]*gjson.Result
	ResourceReferences map[string]*gjson.Result
	DataResources      map[string]resources.DataResource
	ProviderConfigs    map[string]string
}
