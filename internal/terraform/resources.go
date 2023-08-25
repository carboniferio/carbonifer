package terraform

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/carboniferio/carbonifer/internal/utils"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"

	log "github.com/sirupsen/logrus"
)

type GeneralConfig struct {
	JSONData  map[string]string `json:"json_data"`
	DiskTypes struct {
		Default string            `json:"default"`
		Types   map[string]string `json:"types"`
	} `json:"disk_types"`
	IgnoredResources []string `json:"ignored_resources"`
}

var GeneralMappingConfig *GeneralConfig
var TfPlan *map[string]interface{}

func GetResources(tfplan *map[string]interface{}) (map[string]resources.Resource, error) {
	TfPlan = tfplan

	// Get resources from Terraform plan
	plannedResourcesResult, err := utils.JsonGet(".planned_values.root_module.resources", *TfPlan)
	if err != nil {
		return nil, err
	}
	if len(plannedResourcesResult) == 0 {
		return nil, errors.New("No resources found in Terraform plan")
	}
	plannedResources := plannedResourcesResult[0].([]interface{})
	log.Debugf("Reading resources from Terraform plan: %d resources", len(plannedResources))
	resourcesMap := map[string]resources.Resource{}

	mappings, err := LoadMapping("internal/terraform/gcp/mappings")
	if err != nil {
		return nil, err
	}
	// Get general config
	generalMapping := mappings.General
	GeneralMappingConfig, err = convertToGeneralConfig(generalMapping)
	if err != nil {
		return nil, err
	}

	// Get compute resources
	for resourceType, mapping := range mappings.ComputeResource {
		resources, err := GetResourcesOfType(resourceType, mapping.(map[string]interface{}))
		if err != nil {
			return nil, err
		}

		for _, resource := range resources {
			resourcesMap[resource.GetAddress()] = resource
		}
	}

	// Get resource not in mapping
	for _, resourceI := range plannedResources {
		resource := resourceI.(map[string]interface{})
		resourceAddress := resource["address"].(string)
		resourceMap := resourcesMap[resourceAddress]
		if resourceMap == nil {
			// That is an unsupported resource
			resourceType := resource["type"].(string)
			if checkIgnoredResource(resourceType) {
				continue
			}
			unsupportedResource := resources.UnsupportedResource{
				Identification: &resources.ResourceIdentification{
					Name:         resource["name"].(string),
					ResourceType: resourceType,
					// TODO get provider from Terraform plan
					Provider: providers.GCP,
					Count:    1,
				},
			}
			resourcesMap[resourceAddress] = unsupportedResource
		}
	}

	return resourcesMap, nil
}

func checkIgnoredResource(resourceType string) bool {
	ignoredResourceNames := GeneralMappingConfig.IgnoredResources
	for _, ignoredResource := range ignoredResourceNames {
		if ignoredResource == resourceType {
			return true
		}
		// Case of regex
		regex := regexp.MustCompile(ignoredResource)
		if regex.MatchString(resourceType) {
			return true
		}

	}
	return false
}

func convertToGeneralConfig(generalMapping map[string]interface{}) (*GeneralConfig, error) {
	var generalConfig GeneralConfig

	// Convert map to JSON
	jsonData, err := json.Marshal(generalMapping)
	if err != nil {
		return nil, err
	}

	// Convert JSON to GeneralConfig struct
	err = json.Unmarshal(jsonData, &generalConfig)
	if err != nil {
		return nil, err
	}

	return &generalConfig, nil
}

func GetResourcesOfType(resourceType string, mapping map[string]interface{}) ([]resources.Resource, error) {
	pathsProperty := mapping["paths"]
	paths, err := ReadPaths(pathsProperty)
	if err != nil {
		return nil, err
	}

	resourcesResult := []resources.Resource{}
	for _, path := range paths {
		log.Debugf("  Reading resources of type '%s' from path '%s'", resourceType, path)
		resourcesFound, err := utils.JsonGet(path, *TfPlan)
		if err != nil {
			return nil, err
		}
		log.Debugf("  Found %d resources of type '%s'", len(resourcesFound), resourceType)
		for _, resourceI := range resourcesFound {
			resourcesResult, err = GetComputeResource(resourceI, mapping, resourcesResult)
			if err != nil {
				return nil, err
			}

		}
	}
	return resourcesResult, nil

}

func GetComputeResource(resourceI interface{}, mapping map[string]interface{}, resourcesResult []resources.Resource) ([]resources.Resource, error) {
	resource := resourceI.(map[string]interface{})
	resourceAddress := resource["address"].(string)
	context := TFContext{
		ResourceAddress: resourceAddress,
		Mapping:         mapping,
		Resource:        resource,
	}
	name, err := GetString("name", context)
	if err != nil {
		return nil, err
	}
	region, err := GetString("region", context)
	if err != nil {
		return nil, err
	}

	// TODO case of region missing (aws)

	resourceType, err := GetString("type", context)
	if err != nil {
		return nil, err
	}

	index := resource["index"]
	if index != nil {
		nameStr := fmt.Sprintf("%s[%d]", *name, int(index.(float64)))
		name = &nameStr
	}

	computeResource := resources.ComputeResource{
		Identification: &resources.ResourceIdentification{
			Name:         *name,
			ResourceType: *resourceType,
			Provider:     providers.GCP,
			Region:       *region,
		},
		Specs: &resources.ComputeResourceSpecs{},
	}

	// Add vCPUs
	vcpus, err := GetValue("vCPUs", context)
	if err != nil {
		return nil, err
	}
	if vcpus != nil && vcpus.Value != nil {

		intValue, err := utils.ParseToInt(vcpus.Value)
		if err != nil {
			return nil, err
		}
		computeResource.Specs.VCPUs = int32(intValue)

	}

	// Add memory
	memory, err := GetValue("memory", context)
	if err != nil {
		return nil, err
	}
	if memory != nil && memory.Value != nil {
		intValue, err := utils.ParseToInt(memory.Value)
		if err != nil {
			return nil, err
		}
		computeResource.Specs.MemoryMb = int32(intValue)
		unit := strings.ToLower(*memory.Unit)
		switch unit {
		case "gb":
			computeResource.Specs.MemoryMb *= 1024
		case "tb":
			computeResource.Specs.MemoryMb *= 1024 * 1024
		case "pb":
			computeResource.Specs.MemoryMb *= 1024 * 1024 * 1024
		case "mb":
			// nothing to do
		case "kb":
			computeResource.Specs.MemoryMb /= 1024
		case "b":
			computeResource.Specs.MemoryMb /= 1024 * 1024
		default:
			log.Fatalf("Unknown unit for memory: %v", unit)
		}
	}

	// Add GPUs
	gpus, err := GetSlice("guest_accelerator", context)
	if err != nil {
		return nil, err
	}
	for _, gpuI := range gpus {
		gpu := gpuI.(map[string]interface{})
		gpuTypes, err := getGPU(gpu)
		if err != nil {
			return nil, errors.Wrapf(err, "Cannot get GPU types for %v", resourceAddress)
		}
		computeResource.Specs.GpuTypes = append(computeResource.Specs.GpuTypes, gpuTypes...)
	}

	// Add CPU type
	cpuType, err := GetString("cpu_platform", context)
	if err != nil {
		return nil, err
	}
	if cpuType != nil {
		computeResource.Specs.CPUType = *cpuType
	}

	// Add replication factor
	replicationFactor, err := GetValue("replication_factor", context)
	if err != nil {
		return nil, err
	}
	if replicationFactor != nil && replicationFactor.Value != nil {
		intValue, err := utils.ParseToInt(replicationFactor.Value)
		if err != nil {
			return nil, err
		}
		computeResource.Specs.ReplicationFactor = int32(intValue)
	}

	// Add count (case of autoscaling group)
	count, err := GetValue("count", context)
	if err != nil {
		return nil, err
	}
	if count != nil && count.Value != nil {
		intValue, err := utils.ParseToInt(count.Value)
		if err != nil {
			return nil, err
		}
		computeResource.Identification.Count = int64(intValue)
	} else {
		computeResource.Identification.Count = 1
	}

	// Add storage
	storages, err := GetSlice("storage", context)
	if err != nil {
		return nil, err
	}

	for _, storageI := range storages {
		storage := getStorage(storageI.(map[string]interface{}))
		size := storage.SizeGb
		if storage.IsSSD {
			computeResource.Specs.SsdStorage = computeResource.Specs.SsdStorage.Add(size)
		} else {
			computeResource.Specs.HddStorage = computeResource.Specs.HddStorage.Add(size)
		}
	}

	resourcesResult = append(resourcesResult, computeResource)
	log.Debugf("    Reading resource '%s'", computeResource.GetAddress())
	return resourcesResult, nil
}

func getGPU(gpu map[string]interface{}) ([]string, error) {
	gpuTypes := []string{}
	gpuType := gpu["type"].(*ValueWithUnit)
	if gpuType == nil {
		return nil, errors.Errorf("Cannot find GPU type")
	}
	count := gpu["count"].(*ValueWithUnit)
	if count != nil && count.Value != nil {
		intValue, err := utils.ParseToInt(count.Value)
		if err != nil {
			return nil, err
		}
		for i := 0; i < intValue; i++ {
			gpuTypeValue := gpuType.Value.(string)
			gpuTypes = append(gpuTypes, gpuTypeValue)
		}
	}
	return gpuTypes, nil
}

func getStorage(storageMap map[string]interface{}) *Storage {
	storageSize := storageMap["size"].(*ValueWithUnit)
	storageSizeGb, err := decimal.NewFromString(fmt.Sprintf("%v", storageSize.Value))
	if err != nil {
		log.Fatal(err)
	}
	storageType := storageMap["type"].(*ValueWithUnit)
	// TODO get storage size unit correctly
	unit := storageSize.Unit
	if unit != nil {
		if strings.ToLower(*unit) == "mb" {
			storageSizeGb = storageSizeGb.Div(decimal.NewFromInt32(1024))
		}
		if strings.ToLower(*unit) == "tb" {
			storageSizeGb = storageSizeGb.Mul(decimal.NewFromInt32(1024))
		}
		if strings.ToLower(*unit) == "kb" {
			storageSizeGb = storageSizeGb.Div(decimal.NewFromInt32(1024)).Div(decimal.NewFromInt32(1024))
		}
		if strings.ToLower(*unit) == "b" {
			storageSizeGb = storageSizeGb.Div(decimal.NewFromInt32(1024)).Div(decimal.NewFromInt32(1024)).Div(decimal.NewFromInt32(1024))
		}
	}

	isSSD := false
	if storageType != nil {
		if strings.ToLower(storageType.Value.(string)) == "ssd" {
			isSSD = true
		}
	}
	storage := Storage{
		SizeGb: storageSizeGb,
		IsSSD:  isSSD,
	}
	return &storage
}

func GetPropertyMappings(key string, context TFContext) []map[string]interface{} {
	resourcePropertiesMapping := GetMappingProperties(context.Mapping)
	mappingPropertyI, ok := resourcePropertiesMapping[key]
	if !ok {
		log.Debugf("Cannot find resource properties mapping %v of resource %v", key, context.ResourceAddress)
		return nil
	}
	var propertyMappings []map[string]interface{}
	propertyMappingsI, ok := mappingPropertyI.([]interface{})
	if !ok {
		mappingPropertyUnique, ok := mappingPropertyI.(map[string]interface{})
		if !ok {
			mappingPropertyUniqueI, ok := mappingPropertyI.(map[interface{}]interface{})
			if !ok {
				log.Fatalf("Cannot find property mapping %v of resource %v", key, context.ResourceAddress)
			}
			var errConv error
			mappingPropertyUnique, errConv = convertMapKeysToStrings(mappingPropertyUniqueI)
			if errConv != nil {
				log.Fatalf("Cannot convert property mapping %v of resource %v: %v", key, context.ResourceAddress, errConv)
			}
		}
		propertyMappings = []map[string]interface{}{mappingPropertyUnique}
	} else {
		var errMapping error
		propertyMappings, errMapping = ConvertInterfaceSlicesToMapSlice(propertyMappingsI)
		if errMapping != nil {
			log.Fatalf("Cannot convert property mapping %v of resource %v: %v", key, context.ResourceAddress, errMapping)
		}
	}
	return propertyMappings
}
