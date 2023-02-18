package gcp

import (
	"strings"

	"github.com/carboniferio/carbonifer/internal/resources"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

func getComputeInstanceGroupManagerSpecs(
	tfResource tfjson.StateResource,
	dataResources *map[string]resources.DataResource,
	resourceTemplates *map[string]*tfjson.StateResource,
	resourceConfigs *map[string]*tfjson.ConfigResource) (*resources.ComputeResourceSpecs, int64) {

	targetSize := int64(0)
	targetSizeExpr := tfResource.AttributeValues["target_size"]
	if targetSizeExpr != nil {
		targetSize = decimal.NewFromFloat(targetSizeExpr.(float64)).BigInt().Int64()
	}

	var template *tfjson.StateResource
	templateConfig := (*resourceConfigs)[tfResource.Address]
	versionExpr := templateConfig.Expressions["version"]
	if versionExpr != nil {
		for _, version := range versionExpr.NestedBlocks {
			instanceTemplate := version["instance_template"]
			if instanceTemplate != nil {
				references := instanceTemplate.References
				for _, reference := range references {
					if !strings.HasSuffix(reference, ".id") {
						template = (*resourceTemplates)[reference]
					}
				}
			}
		}
	}

	if template != nil {
		zone := tfResource.AttributeValues["zone"]
		if zone == nil {
			log.Fatalf("No zone declared for %v", tfResource.Address)
		}
		templateResource := GetResourceTemplate(*template, dataResources, zone.(string))
		computeTemplate, ok := templateResource.(resources.ComputeResource)
		if ok {
			return computeTemplate.Specs, targetSize
		} else {
			log.Fatalf("Type mismatch, not a esources.ComputeResource template %v", computeTemplate.GetAddress())
		}
	}
	return nil, 0
}
