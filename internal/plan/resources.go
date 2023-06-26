package plan

import (
	"encoding/json"
	"strings"

	"github.com/carboniferio/carbonifer/internal/plan/aws"
	"github.com/carboniferio/carbonifer/internal/plan/gcp"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/carboniferio/carbonifer/internal/terraform/tfrefs"
	tfjson "github.com/hashicorp/terraform-json"

	log "github.com/sirupsen/logrus"
)

func GetResources(tfPlan *tfjson.Plan) (map[string]resources.Resource, error) {

	log.Debugf("Reading resources from Terraform plan: %d resources", len(tfPlan.PlannedValues.RootModule.Resources))
	resourcesMap := make(map[string]resources.Resource)
	terraformRefs := tfrefs.References{
		ResourceConfigs:    map[string]*tfjson.ConfigResource{},
		ResourceReferences: map[string]*tfjson.StateResource{},
		DataResources:      map[string]resources.DataResource{},
		ProviderConfigs:    map[string]string{},
	}
	var planDataRes = tfPlan.PlannedValues.RootModule.Resources
	if tfPlan.PriorState != nil {
		planDataRes = tfPlan.PriorState.Values.RootModule.Resources
	}
	for _, priorRes := range planDataRes {
		log.Debugf("Reading prior state resources %v", priorRes.Address)
		if priorRes.Mode == "data" {
			if strings.HasPrefix(priorRes.Type, "google") {
				dataResource := gcp.GetDataResource(*priorRes)
				terraformRefs.DataResources[dataResource.GetKey()] = dataResource
			}
			if strings.HasPrefix(priorRes.Type, "aws") {
				dataResource := aws.GetDataResource(*priorRes)
				terraformRefs.DataResources[dataResource.GetKey()] = dataResource
			}
		}
	}

	// Find template first
	for _, res := range tfPlan.PlannedValues.RootModule.Resources {
		log.Debugf("Reading resource %v", res.Address)
		if strings.HasPrefix(res.Type, "google") && (strings.HasSuffix(res.Type, "_template") ||
			strings.HasSuffix(res.Type, "_autoscaler")) {
			if res.Mode == "managed" {
				terraformRefs.ResourceReferences[res.Address] = res
			}
		}
	}

	// Index configurations in order to find relationships
	for _, resConfig := range tfPlan.Config.RootModule.Resources {
		log.Debugf("Reading resource config %v", resConfig.Address)
		if strings.HasPrefix(resConfig.Type, "google") {
			if resConfig.Mode == "managed" {
				terraformRefs.ResourceConfigs[resConfig.Address] = resConfig
			}
		}
	}

	// Get default values
	for provider, resConfig := range tfPlan.Config.ProviderConfigs {
		if provider == "aws" {
			aws.GetDefaults(resConfig, tfPlan, &terraformRefs)
		}
	}

	// Get All resources
	for _, res := range tfPlan.PlannedValues.RootModule.Resources {
		log.Debugf("Reading resource %v", res.Address)

		if res.Mode == "managed" {
			var resource resources.Resource
			prefix := strings.Split(res.Type, "_")[0]
			if prefix == "google" {
				resource = gcp.GetResource(*res, &terraformRefs)
			} else if prefix == "aws" {
				resource = aws.GetResource(*res, &terraformRefs)
			} else {
				log.Warnf("Skipping resource %s. Provider not supported : %s", res.Type, prefix)
			}
			if resource != nil {
				resourcesMap[resource.GetAddress()] = resource
				if log.IsLevelEnabled(log.DebugLevel) {
					computeJsonStr := "<RESOURCE TYPE CURRENTLY NOT SUPPORTED>"
					if resource.IsSupported() {
						computeJson, _ := json.Marshal(resource)
						computeJsonStr = string(computeJson)
					}
					log.Debugf("  Compute resource : %v", string(computeJsonStr))
				}
			}
		}

	}
	return resourcesMap, nil
}
