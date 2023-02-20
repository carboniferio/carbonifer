package gcp

import (
	"strings"

	"github.com/carboniferio/carbonifer/internal/resources"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func getComputeInstanceGroupManagerSpecs(
	tfResource tfjson.StateResource,
	dataResources *map[string]resources.DataResource,
	resourceReferences *map[string]*tfjson.StateResource,
	resourceConfigs *map[string]*tfjson.ConfigResource) (*resources.ComputeResourceSpecs, int64) {

	// Get template of instance
	specs, targetSize := getGroupInstanceTemplateSpecs(tfResource, dataResources, resourceReferences, resourceConfigs)
	if specs == nil {
		return specs, targetSize
	}

	// Get targetSize from autoscaler if exists
	var autoscaler *tfjson.StateResource
	for _, resourceConfig := range *resourceConfigs {
		if resourceConfig.Type == "google_compute_autoscaler" {
			targetExpr := (*resourceConfig).Expressions["target"]
			if targetExpr != nil {
				for _, target := range (*targetExpr).References {
					if target == tfResource.Address {
						autoscaler = (*resourceReferences)[resourceConfig.Address]
						break
					}
				}
				if autoscaler != nil {
					break
				}
			}
		}
	}
	if autoscaler != nil {
		targetSize = getTargetSizeFromAutoscaler(autoscaler, resourceConfigs, tfResource, resourceReferences, targetSize)
	}

	return specs, targetSize
}

func getTargetSizeFromAutoscaler(autoscaler *tfjson.StateResource, resourceConfigs *map[string]*tfjson.ConfigResource, tfResource tfjson.StateResource, resourceReferences *map[string]*tfjson.StateResource, targetSizeOfTemplate int64) int64 {

	targetSize := targetSizeOfTemplate
	autoscalingPoliciesI := autoscaler.AttributeValues["autoscaling_policy"]
	if autoscalingPoliciesI != nil {
		for _, autoscalingPolicyI := range autoscalingPoliciesI.([]interface{}) {
			autoscalingPolicy := autoscalingPolicyI.(map[string]interface{})
			minSize := autoscalingPolicy["min_replicas"]
			if minSize == nil {
				minSize = 0
			}
			maxSize := autoscalingPolicy["max_replicas"]
			if maxSize == nil {
				maxSize = 0
			}
			targetSize = computeTargetSize(decimal.NewFromFloat(minSize.(float64)), decimal.NewFromFloat(maxSize.(float64)))
		}
	}

	return targetSize
}

func computeTargetSize(minSize decimal.Decimal, maxSize decimal.Decimal) int64 {
	avgAutoscalerSizePercent := decimal.NewFromFloat(viper.GetFloat64("provider.gcp.avg_autoscaler_size_percent"))
	return avgAutoscalerSizePercent.Mul(maxSize.Sub(minSize)).Ceil().IntPart()
}

func getGroupInstanceTemplateSpecs(
	tfResource tfjson.StateResource,
	dataResources *map[string]resources.DataResource,
	resourceReferences *map[string]*tfjson.StateResource,
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
						template = (*resourceReferences)[reference]
					}
				}
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
		templateResource := GetResourceTemplate(*template, dataResources, zones[0])
		computeTemplate, ok := templateResource.(resources.ComputeResource)
		if ok {
			return computeTemplate.Specs, targetSize
		} else {
			log.Fatalf("Type mismatch, not a esources.ComputeResource template %v", computeTemplate.GetAddress())
		}
	}
	return nil, 0
}
