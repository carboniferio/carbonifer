package gcp

import (
	"strings"

	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/carboniferio/carbonifer/internal/terraform/tfrefs"
	tfjson "github.com/hashicorp/terraform-json"
	log "github.com/sirupsen/logrus"
)

func getComputeResourceFromTemplateSpecs(
	tfResource tfjson.StateResource,
	tfRefs *tfrefs.References) *resources.ComputeResourceSpecs {

	// Get template of instance
	specs := getTemplateSpecs(tfResource, tfRefs)
	if specs != nil {
		return specs
	}
	return nil

}

func getTemplateSpecs(
	tfResource tfjson.StateResource,
	tfRefs *tfrefs.References) *resources.ComputeResourceSpecs {

	// Find google_compute_instance_from_template resourceConfig
	iftConfig := (tfRefs.ResourceConfigs)[tfResource.Address]

	var template *tfjson.StateResource
	sourceTemplateExpr := iftConfig.Expressions["source_instance_template"]
	if sourceTemplateExpr != nil {
		references := sourceTemplateExpr.References
		for _, reference := range references {
			if !strings.HasSuffix(reference, ".id") {
				template = (tfRefs.ResourceReferences)[reference]
				break
			}
		}
	}

	if template != nil {
		var zones []string
		zoneAttr := tfResource.AttributeValues["zone"]
		if zoneAttr != nil {
			zones = append(zones, zoneAttr.(string))
		}
		distributionPolicyZonesI := tfResource.AttributeValues["distribution_policy_zones"]
		if distributionPolicyZonesI != nil {
			distributionPolicyZones := distributionPolicyZonesI.([]interface{})
			for _, z := range distributionPolicyZones {
				zones = append(zones, z.(string))
			}
		}

		if len(zones) == 0 {
			log.Fatalf("No zone or distribution policy declared for %v", tfResource.Address)
		}
		templateResource := GetResourceTemplate(*template, tfRefs, zones[0])
		computeTemplate, ok := templateResource.(resources.ComputeResource)
		if ok {
			return computeTemplate.Specs
		} else {
			log.Fatalf("Type mismatch, not a esources.ComputeResource template %v", computeTemplate.GetAddress())
		}
	} else {
		log.Fatalf("Cannot find template of %v", tfResource.Address)
	}
	return nil
}
