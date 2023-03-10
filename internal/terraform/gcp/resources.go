package gcp

import (
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/forestgiant/sliceutil"
	tfjson "github.com/hashicorp/terraform-json"
)

func GetResource(
	tfResource tfjson.StateResource,
	dataResources *map[string]resources.DataResource,
	resourceTemplates *map[string]*tfjson.StateResource,
	resourceConfigs *map[string]*tfjson.ConfigResource) resources.Resource {

	resourceId := getResourceIdentification(tfResource)
	if resourceId.ResourceType == "google_compute_instance" {
		specs := getComputeResourceSpecs(tfResource, dataResources, nil)
		return resources.ComputeResource{
			Identification: resourceId,
			Specs:          specs,
		}
	}
	if resourceId.ResourceType == "google_compute_instance_from_template" {
		specs := getComputeResourceFromTemplateSpecs(tfResource, dataResources, resourceTemplates, resourceConfigs)
		if specs != nil {
			return resources.ComputeResource{
				Identification: resourceId,
				Specs:          specs,
			}
		}
	}
	if resourceId.ResourceType == "google_compute_disk" ||
		resourceId.ResourceType == "google_compute_region_disk" {
		specs := getComputeDiskResourceSpecs(tfResource, dataResources)
		return resources.ComputeResource{
			Identification: resourceId,
			Specs:          specs,
		}
	}
	if resourceId.ResourceType == "google_sql_database_instance" {
		specs := getSQLResourceSpecs(tfResource)
		return resources.ComputeResource{
			Identification: resourceId,
			Specs:          specs,
		}
	}
	if resourceId.ResourceType == "google_compute_instance_group_manager" ||
		resourceId.ResourceType == "google_compute_region_instance_group_manager" {
		specs, count := getComputeInstanceGroupManagerSpecs(tfResource, dataResources, resourceTemplates, resourceConfigs)
		if specs != nil {
			resourceId.Count = count
			return resources.ComputeResource{
				Identification: resourceId,
				Specs:          specs,
			}
		}
	}
	ignoredResourceType := []string{
		"google_compute_autoscaler",
		"google_compute_instance_template",
	}
	if sliceutil.Contains(ignoredResourceType, resourceId.ResourceType) {
		return nil
	}

	return resources.UnsupportedResource{
		Identification: resourceId,
	}
}

func GetResourceTemplate(tfResource tfjson.StateResource, dataResources *map[string]resources.DataResource, zone string) resources.Resource {
	resourceId := getResourceIdentification(tfResource)
	if resourceId.ResourceType == "google_compute_instance_template" {
		specs := getComputeResourceSpecs(tfResource, dataResources, zone)
		return resources.ComputeResource{
			Identification: resourceId,
			Specs:          specs,
		}
	}
	return nil
}
