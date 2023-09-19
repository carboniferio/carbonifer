package plan

import (
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

// TfPlan is the Terraform plan
var TfPlan *map[string]interface{}

// GetResources returns the resources of the Terraform plan
func GetResources(tfplan *map[string]interface{}) (map[string]resources.Resource, error) {
	TfPlan = tfplan

	plannedResources := []interface{}{}

	// Get resources from Terraform plan
	jqPath := ".planned_values | .. | objects | select(has(\"resources\")) | .resources[]"
	plannedResourcesResult, err := utils.GetJSON(jqPath, *TfPlan)

	if err != nil {
		return nil, err
	}
	if len(plannedResourcesResult) > 0 {
		for _, plannedResourceI := range plannedResourcesResult {
			plannedResource, ok := plannedResourceI.(map[string]interface{})
			if !ok {
				return nil, errors.Errorf("Cannot parse planned resource %v", plannedResourceI)
			}
			plannedResources = append(plannedResources, plannedResource)
		}
	}

	// Get compute resources
	resourcesMap := map[string]resources.Resource{}
	mapping, err := GetMapping()
	if err != nil {
		errW := errors.Wrap(err, "Cannot get mapping")
		return nil, errW
	}
	for resourceType, mapping := range *mapping.ComputeResource {
		resources, err := getResourcesOfType(resourceType, &mapping)
		if err != nil {
			errW := errors.Wrapf(err, "Cannot get resources of type %v", resourceType)
			return nil, errW
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
		providerName := resource["provider_name"].(string)
		provider, err := parseProvider(providerName)
		if err != nil {
			continue
		}
		if resourceMap == nil {
			// That is an unsupported resource
			resourceType := resource["type"].(string)
			if checkIgnoredResource(resourceType, provider) {
				continue
			}
			unsupportedResource := resources.UnsupportedResource{
				Identification: &resources.ResourceIdentification{
					Address:      resourceAddress,
					Name:         resource["name"].(string),
					ResourceType: resourceType,
					Provider:     provider,
					Count:        1,
				},
			}
			resourcesMap[resourceAddress] = unsupportedResource
		}
	}

	return resourcesMap, nil
}

func checkIgnoredResource(resourceType string, provider providers.Provider) bool {
	ignoredResourceNames := (*globalMappings.General)[provider].IgnoredResources
	if ignoredResourceNames != nil {
		for _, ignoredResource := range *ignoredResourceNames {
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
	return false
}
func getResourcesOfType(resourceType string, mapping *ResourceMapping) ([]resources.Resource, error) {
	pathsProperty := mapping.Paths
	paths, err := readPaths(pathsProperty)
	if err != nil {
		errW := errors.Wrapf(err, "Cannot read paths of resource type %v", resourceType)
		return nil, errW
	}

	resourcesResult := []resources.Resource{}
	for _, path := range paths {
		log.Debugf("  Reading resources of type '%s' from path '%s'", resourceType, path)
		resourcesFound, err := getJSON(path, *TfPlan)
		if err != nil {
			errW := errors.Wrapf(err, "Cannot find resource for path %v", path)
			return nil, errW
		}
		log.Debugf("  Found %d resources of type '%s'", len(resourcesFound), resourceType)
		for _, resourceI := range resourcesFound {
			resourcesResultGot, err := GetComputeResource(resourceI, mapping, resourcesResult)
			if err != nil {
				errW := errors.Wrapf(err, "Cannot get compute resource for path %v", path)
				return nil, errW
			}
			if resourcesResultGot != nil {
				resourcesResult = resourcesResultGot
			}

		}
	}
	return resourcesResult, nil

}

func GetComputeResource(resourceI interface{}, resourceMapping *ResourceMapping, resourcesResult []resources.Resource) ([]resources.Resource, error) {
	resource := resourceI.(map[string]interface{})
	resourceAddress := resource["address"].(string)
	providerName, ok := resource["provider_name"].(string)
	if !ok {
		return nil, errors.Errorf("Cannot find provider name for resource %v", resourceAddress)
	}
	provider, err := parseProvider(providerName)
	if err != nil {
		return nil, nil
	}
	contextObject := tfContext{
		ResourceAddress: resourceAddress,
		Mapping:         resourceMapping,
		Resource:        resource,
		Provider:        provider,
	}
	contextObject.RootContext = &contextObject
	context := &contextObject
	name, err := getString("name", context)
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot get name for resource %v", resourceAddress)
	}
	region, err := getString("region", context)
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot get region for resource %v", resourceAddress)
	}
	if region == nil {
		region = getDefaultRegion()
		if region == nil {
			return nil, errors.Errorf("Cannot find default region for resource %v", resourceAddress)
		}
	}

	resourceType, err := getString("type", context)
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot get type for resource %v", resourceAddress)
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
			Provider:     provider,
			Region:       *region,
			Address:      resourceAddress,
		},
		Specs: &resources.ComputeResourceSpecs{
			HddStorage: decimal.Zero,
			SsdStorage: decimal.Zero,
		},
	}

	// Add vCPUs
	vcpus, err := getValue("vCPUs", context)
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot get vCPUs for %v", resourceAddress)
	}
	if vcpus != nil && vcpus.Value != nil {

		intValue, err := utils.ParseToInt(vcpus.Value)
		if err != nil {
			return nil, errors.Wrapf(err, "Cannot parse vCPUs for %v", resourceAddress)
		}
		computeResource.Specs.VCPUs = int32(intValue)

	}

	// Add memory
	memory, err := getValue("memory", context)
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot get memory for %v", resourceAddress)
	}
	if memory != nil && memory.Value != nil {
		intValue, err := utils.ParseToInt(memory.Value)
		if err != nil {
			return nil, errors.Wrapf(err, "Cannot parse memory for %v", resourceAddress)
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
	gpus, err := getSlice("guest_accelerator", context)
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot get GPUs for %v", resourceAddress)
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
	cpuType, err := getString("cpu_platform", context)
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot get CPU platform for %v", resourceAddress)
	}
	if cpuType != nil {
		computeResource.Specs.CPUType = *cpuType
	}

	// Add replication factor
	replicationFactor, err := getValue("replication_factor", context)
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot get replication factor for %v", resourceAddress)
	}
	if replicationFactor != nil && replicationFactor.Value != nil {
		intValue, err := utils.ParseToInt(replicationFactor.Value)
		if err != nil {
			return nil, errors.Wrapf(err, "Cannot parse replication factor for %v", resourceAddress)
		}
		computeResource.Identification.ReplicationFactor = int32(intValue)
	} else {
		computeResource.Identification.ReplicationFactor = 1
	}

	// Add count (case of autoscaling group)
	count, err := getValue("count", context)
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot get count for %v", resourceAddress)
	}
	if count != nil && count.Value != nil {
		intValue, err := utils.ParseToInt(count.Value)
		if err != nil {
			return nil, errors.Wrapf(err, "Cannot parse count for %v", resourceAddress)
		}
		computeResource.Identification.Count = int64(intValue)
	} else {
		computeResource.Identification.Count = 1
	}

	// Add storage
	storages, err := getSlice("storage", context)
	if err != nil {
		return nil, errors.Wrapf(err, "Cannot get storages for %v", resourceAddress)
	}

	for i, storageI := range storages {
		if storageI == nil {
			continue
		}
		storage, err := getStorage(storageI.(map[string]interface{}))
		if err != nil {
			return nil, errors.Wrapf(err, "Cannot get storage[%v] for %v", i, resourceAddress)
		}
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
	gpuType := gpu["type"].(*valueWithUnit)
	if gpuType == nil {
		return nil, errors.Errorf("Cannot find GPU type")
	}
	count := gpu["count"].(*valueWithUnit)
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

func getStorage(storageMap map[string]interface{}) (*storage, error) {
	storageSize, ok := storageMap["size"].(*valueWithUnit)
	if !ok {
		// It can happen there is no size but type as been set by default. In this case, we ignore the storage
		// Can be fixed in the mapping by selecting only if the size property is set (example '.values | select(.allocated_storage)')
		log.Warnf("Cannot find storage size in storageMap '%v': %T", storageMap, storageMap)
		return nil, nil
	}
	if storageSize == nil {
		return nil, errors.Errorf("Storage size is nil '%v': %T", storageSize, storageSize)
	}
	if storageSize.Value == nil {
		return nil, errors.Errorf("Storage size value is nil '%v': %T", storageSize, storageSize)
	}
	storageSizeGb, err := decimal.NewFromString(fmt.Sprintf("%v", storageSize.Value))
	if err != nil {
		log.Fatal(err)
	}
	storageType := storageMap["type"].(*valueWithUnit)
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
		diskType, ok := storageType.Value.(*DiskType)
		if !ok {
			diskTypeStr, ok := storageType.Value.(string)
			if !ok {
				return nil, errors.Errorf("Cannot find storage type '%v': %T", storageType.Value, storageType.Value)
			}
			diskTypeParsed, err := ParseDiskType(diskTypeStr)
			if err != nil {
				return nil, errors.Wrapf(err, "Cannot parse disk type '%v'", diskTypeStr)
			}
			diskType = &diskTypeParsed
		}
		if *diskType == SSD {
			isSSD = true
		}
	}
	storage := storage{
		SizeGb: storageSizeGb,
		IsSSD:  isSSD,
	}
	return &storage, nil
}

func parseProvider(tfProviderName string) (providers.Provider, error) {
	if strings.HasSuffix(tfProviderName, "google") {
		return providers.ParseProvider("gcp")
	}
	if strings.HasSuffix(tfProviderName, "aws") {
		return providers.ParseProvider("aws")
	}
	return providers.ParseProvider(tfProviderName)
}
