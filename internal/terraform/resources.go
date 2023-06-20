package terraform

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/carboniferio/carbonifer/internal/terraform/aws"
	"github.com/carboniferio/carbonifer/internal/terraform/gcp"
	"github.com/carboniferio/carbonifer/internal/terraform/tfrefs"
	"github.com/tidwall/gjson"

	log "github.com/sirupsen/logrus"
)

func GetResources(tfPlan *string) (map[string]resources.Resource, error) {

	tfPlanJson := gjson.Parse(*tfPlan)
	plannedResources := tfPlanJson.Get("planned_values.root_module.resources").Array()

	log.Debugf("Reading resources from Terraform plan: %d resources", len(plannedResources))
	resourcesMap := make(map[string]resources.Resource)
	terraformRefs := tfrefs.References{
		ResourceConfigs:    map[string]*gjson.Result{},
		ResourceReferences: map[string]*gjson.Result{},
		DataResources:      map[string]resources.DataResource{},
		ProviderConfigs:    map[string]string{},
	}
	planDataRes := plannedResources
	if tfPlanJson.Get("prior_state").Exists() {
		planDataRes = append(planDataRes, tfPlanJson.Get("prior_state.root_module.resources").Array()...)
	}
	for _, priorRes := range planDataRes {
		log.Debugf("Reading prior state resources %v", priorRes.Get("address").String())
		if priorRes.Get("mode").String() == "data" {
			resType := priorRes.Get("type").String()
			if strings.HasPrefix(resType, "google") {
				dataResource := gcp.GetDataResource(&priorRes)
				terraformRefs.DataResources[dataResource.GetKey()] = dataResource
			}
			if strings.HasPrefix(resType, "aws") {
				dataResource := aws.GetDataResource(&priorRes)
				terraformRefs.DataResources[dataResource.GetKey()] = dataResource
			}
		}
	}

	// Find template first
	for _, res := range plannedResources {
		resAddress := res.Get("address").String()
		log.Debugf("Reading resource %v", resAddress)
		resType := res.Get("type").String()
		if strings.HasPrefix(resType, "google") && (strings.HasSuffix(resType, "_template") ||
			strings.HasSuffix(resType, "_autoscaler")) {
			if res.Get("mode").String() == "managed" {
				terraformRefs.ResourceReferences[resAddress] = &res
			}
		}
	}

	// Index configurations in order to find relationships
	for _, resConfig := range tfPlanJson.Get("configuration.root_module.resources").Array() {
		resAddress := resConfig.Get("address").String()
		resType := resConfig.Get("type").String()
		log.Debugf("Reading resource config %v", resAddress)
		if strings.HasPrefix(resType, "google") {
			if resConfig.Get("mode").String() == "managed" {
				terraformRefs.ResourceConfigs[resAddress] = &resConfig
			}
		}
	}

	// Get default values
	for provider, resConfig := range tfPlanJson.Get("configuration.provider_config").Map() {
		if provider == "aws" {
			log.Debugf("Reading provider config %v", resConfig.Get("name").String())
			// TODO #58 Improve way we get default regions (env var, profile...)
			region := resConfig.Get("aws.expressions.region.constant_value").String()
			if region == "" {
				if os.Getenv("AWS_REGION") != "" {
					region = os.Getenv("AWS_REGION")
				}
			}
			if region != "" {
				terraformRefs.ProviderConfigs["region"] = region
			}
		}
	}

	// Get All resources
	for _, res := range plannedResources {
		log.Debugf("Reading resource %v", res.Get("address").String())

		if res.Get("mode").String() == "managed" {
			var resource resources.Resource
			prefix := strings.Split(res.Get("type").String(), "_")[0]
			if prefix == "google" {
				resource = gcp.GetResource(&res, &terraformRefs)
			} else if prefix == "aws" {
				resource = aws.GetResource(&res, &terraformRefs)
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
