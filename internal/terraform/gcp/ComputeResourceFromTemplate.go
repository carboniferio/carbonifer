package gcp

import (
	"strings"

	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/carboniferio/carbonifer/internal/terraform/tfrefs"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func getComputeResourceFromTemplateSpecs(
	tfResource *gjson.Result,
	tfRefs *tfrefs.References) *resources.ComputeResourceSpecs {

	// Get template of instance
	specs := getTemplateSpecs(tfResource, tfRefs)
	if specs != nil {
		return specs
	}
	return nil

}

func getTemplateSpecs(
	tfResource *gjson.Result,
	tfRefs *tfrefs.References) *resources.ComputeResourceSpecs {

	// Find google_compute_instance_from_template resourceConfig
	iftConfig := (tfRefs.ResourceConfigs)[tfResource.Get("address").String()]

	var template *gjson.Result
	sourceTemplates := iftConfig.Get("expressions.source_instance_template.references")
	sourceTemplates.ForEach(func(_, ref gjson.Result) bool {
		if strings.HasSuffix(ref.String(), ".id") {
			template = tfRefs.ResourceReferences[ref.String()]
			return false // Stop iterating
		}
		return true // Continue iterating
	})

	if template != nil {

		zones := GetZones(tfResource)

		if len(zones) == 0 {
			log.Fatalf("No zone or distribution policy declared for %v", tfResource.Get("address").String())
		}
		templateResource := GetResourceTemplate(template, tfRefs, zones[0])
		computeTemplate, ok := templateResource.(resources.ComputeResource)
		if ok {
			return computeTemplate.Specs
		} else {
			log.Fatalf("Type mismatch, not a esources.ComputeResource template %v", computeTemplate.GetAddress())
		}
	} else {
		log.Fatalf("Cannot find template of %v", tfResource.Get("address").String())
	}
	return nil
}
