package terraform

import (
	"encoding/json"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"

	log "github.com/sirupsen/logrus"
)

type GeneralConfig struct {
	JSONData  map[string]string `json:"json_data"`
	DiskTypes struct {
		Default string            `json:"default"`
		Types   map[string]string `json:"types"`
	} `json:"disk_types"`
}

var GeneralMappingConfig *GeneralConfig
var TfPlan *map[string]interface{}

func GetResources(tfplan *map[string]interface{}) (map[string]resources.Resource, error) {
	TfPlan = tfplan

	// Get resources from Terraform plan
	plannedResourcesRaw, err := jsonpath.Get("$.planned_values.root_module.resources", *TfPlan)
	if err != nil {
		return nil, err
	}
	plannedResources := plannedResourcesRaw.([]interface{})

	log.Debugf("Reading resources from Terraform plan: %d resources", len(plannedResources))
	resourcesMap := make(map[string]resources.Resource)

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

	return resourcesMap, nil
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
	paths, err := ReadPaths(resourceType, pathsProperty)
	if err != nil {
		return nil, err
	}

	resourcesResult := []resources.Resource{}
	for _, path := range paths {
		log.Debugf("  Reading resources of type '%s' from path '%s'", resourceType, path)
		resourcesRaw, err := jsonpath.Get(path.(string), *TfPlan)
		if err != nil {
			return nil, err
		}
		resourcesFound := resourcesRaw.([]interface{})
		log.Debugf("  Found %d resources of type '%s'", len(resourcesFound), resourceType)
		for _, resourceI := range resourcesFound {
			resourcesResult, err = GetComputeResource(resourceI, resourceType, mapping, resourcesResult)
			if err != nil {
				return nil, err
			}

		}
	}
	return resourcesResult, nil

}

func GetComputeResource(resourceI interface{}, resourceType string, mapping map[string]interface{}, resourcesResult []resources.Resource) ([]resources.Resource, error) {
	resourceAdress := resourceI.(map[string]interface{})["address"].(string)
	resource := resourceI.(map[string]interface{})
	name, err := GetString("name", resourceAdress, resource, mapping)
	if err != nil {
		return nil, err
	}
	region, err := GetString("region", resourceAdress, resource, mapping)
	if err != nil {
		return nil, err
	}

	computeResource := resources.ComputeResource{
		Identification: &resources.ResourceIdentification{
			Name:         *name,
			ResourceType: resourceAdress,
			Provider:     providers.GCP,
			Region:       *region,
		},
		Specs: &resources.ComputeResourceSpecs{},
	}
	vcpus, err := GetValue("vCPUs", resourceAdress, resource, mapping)
	if err != nil {
		return nil, err
	}
	if vcpus != nil && vcpus.Value != nil {
		computeResource.Specs.VCPUs = int32(vcpus.Value.(float64))
	}
	memory, err := GetValue("memory", resourceAdress, resource, mapping)
	if err != nil {
		return nil, err
	}
	if memory != nil && memory.Value != nil {
		computeResource.Specs.MemoryMb = int32(memory.Value.(float64))
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
	// TODO: add GPU
	// TODO: add CPU type
	// TODO: add replication factor

	storages, err := GetSlice("storage", resourceAdress, resource, mapping)
	if err != nil {
		return nil, err
	}

	for _, storageI := range storages {
		storage := storageI.(Storage)
		size := storage.SizeGb
		if storage.IsSSD {
			computeResource.Specs.SsdStorage = computeResource.Specs.SsdStorage.Add(size)
		} else {
			computeResource.Specs.SsdStorage = computeResource.Specs.HddStorage.Add(size)
		}
	}

	resourcesResult = append(resourcesResult, computeResource)
	log.Debugf("    Reading resource '%s'", computeResource.GetAddress())
	return resourcesResult, nil
}

func GetPropertyMappings(mapping map[string]interface{}, key string, resourceType string) []map[string]interface{} {
	resourcePropertiesMapping := GetMappingProperties(mapping)
	mappingPropertyI, ok := resourcePropertiesMapping[key]
	if !ok {
		log.Debugf("Cannot find resource properties mapping %v of resource type %v", key, resourceType)
		return nil
	}
	var propertyMappings []map[string]interface{}
	propertyMappingsI, ok := mappingPropertyI.([]interface{})
	if !ok {
		mappingPropertyUnique, ok := mappingPropertyI.(map[string]interface{})
		if !ok {
			mappingPropertyUniqueI, ok := mappingPropertyI.(map[interface{}]interface{})
			if !ok {
				log.Fatalf("Cannot find property mapping %v of resource type %v", key, resourceType)
			}
			var errConv error
			mappingPropertyUnique, errConv = convertMapKeysToStrings(mappingPropertyUniqueI)
			if errConv != nil {
				log.Fatalf("Cannot convert property mapping %v of resource type %v: %v", key, resourceType, errConv)
			}
		}
		propertyMappings = []map[string]interface{}{mappingPropertyUnique}
	} else {
		var errMapping error
		propertyMappings, errMapping = ConvertInterfaceSlicesToMapSlice(propertyMappingsI)
		if errMapping != nil {
			log.Fatalf("Cannot convert property mapping %v of resource type %v: %v", key, resourceType, errMapping)
		}
	}
	return propertyMappings
}
