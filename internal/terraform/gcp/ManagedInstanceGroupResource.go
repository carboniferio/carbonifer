package gcp

import (
	"strings"

	"github.com/carboniferio/carbonifer/internal/resources"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

func getComputeInstanceGroupManagerSpecs(tfResource tfjson.ConfigResource, dataResources *map[string]resources.DataResource, resourceTemplates *map[string]*tfjson.ConfigResource) (*resources.ComputeResourceSpecs, int64) {
	targetSize := int64(0)
	targetSizeExpr := GetConstFromConfig(&tfResource, "target_size")
	if targetSizeExpr != nil {
		targetSize = decimal.NewFromFloat(targetSizeExpr.(float64)).BigInt().Int64()
	}
	versionExpr := tfResource.Expressions["version"]
	var template *tfjson.ConfigResource
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
		zone := GetConstFromConfig(&tfResource, "zone").(string)
		templateResource := GetResourceTemplate(*template, dataResources, zone)
		computeTemplate, ok := templateResource.(resources.ComputeResource)
		if ok {
			return computeTemplate.Specs, targetSize
		} else {
			log.Fatalf("Type mismatch, not a esources.ComputeResource template %v", computeTemplate.GetAddress())
		}
	}
	return nil, 0
}
