package gcp

import (
	"fmt"
	"strings"

	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/carboniferio/carbonifer/internal/terraform/tfrefs"
	"github.com/carboniferio/carbonifer/internal/utils"
	"github.com/forestgiant/sliceutil"
	"github.com/tidwall/gjson"
)

func GetResource(
	tfResource *gjson.Result,
	tfRefs *tfrefs.References) resources.Resource {

	resourceId := getResourceIdentification(tfResource)
	if resourceId.ResourceType == "google_compute_instance" {
		specs := getComputeResourceSpecs(tfResource, tfRefs, nil)
		return resources.ComputeResource{
			Identification: resourceId,
			Specs:          specs,
		}
	}
	if resourceId.ResourceType == "google_compute_instance_from_template" {
		specs := getComputeResourceFromTemplateSpecs(tfResource, tfRefs)
		if specs != nil {
			return resources.ComputeResource{
				Identification: resourceId,
				Specs:          specs,
			}
		}
	}
	if resourceId.ResourceType == "google_compute_disk" ||
		resourceId.ResourceType == "google_compute_region_disk" {
		specs := getComputeDiskResourceSpecs(tfResource, tfRefs)
		return resources.ComputeResource{
			Identification: resourceId,
			Specs:          specs,
		}
	}

	// TODO
	// if resourceId.ResourceType == "google_sql_database_instance" {
	// 	specs := getSQLResourceSpecs(tfResource)
	// 	return resources.ComputeResource{
	// 		Identification: resourceId,
	// 		Specs:          specs,
	// 	}
	// }

	// if resourceId.ResourceType == "google_compute_instance_group_manager" ||
	// 	resourceId.ResourceType == "google_compute_region_instance_group_manager" {
	// 	specs, count := getComputeInstanceGroupManagerSpecs(tfResource, tfRefs)
	// 	if specs != nil {
	// 		resourceId.Count = count
	// 		return resources.ComputeResource{
	// 			Identification: resourceId,
	// 			Specs:          specs,
	// 		}
	// 	}
	// }
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

func GetResourceTemplate(tfResource *gjson.Result, tfRefs *tfrefs.References, zone string) resources.Resource {
	resourceId := getResourceIdentification(tfResource)
	if resourceId.ResourceType == "google_compute_instance_template" {
		specs := getComputeResourceSpecs(tfResource, tfRefs, &zone)
		return resources.ComputeResource{
			Identification: resourceId,
			Specs:          specs,
		}
	}
	return nil
}

func GetZones(resource *gjson.Result) []string {
	zones := []string{}
	zoneRes := utils.GetOr(resource, []string{"region", "zone", "replica_zones", "distribution_policy_zones"})
	if zoneRes.IsArray() {
		zoneRes.ForEach(func(_, ref gjson.Result) bool {
			zones = append(zones, ref.String())
			return true
		})
	} else {
		zones = append(zones, zoneRes.String())
	}
	return zones
}

func GetRegion(resource *gjson.Result) string {
	region := ""
	fmt.Println(resource)
	regionRes := utils.GetOr(resource, []string{"region", "zone", "replica_zones", "distribution_policy_zones"})
	if regionRes.IsArray() {
		// If the result is an array, get its first item
		region = regionRes.Array()[0].String()
	} else {
		region = regionRes.String()
	}
	parts := strings.Split(region, "-")
	if len(parts) > 1 {
		region = strings.Join(parts[:2], "-")
	}
	return region
}
