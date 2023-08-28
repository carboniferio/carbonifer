package terraform

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/ghodss/yaml"
)

type MappingsFromFiles struct {
	General         map[string]interface{} `yaml:"general"`
	ComputeResource map[string]interface{} `yaml:"compute_resource"`
}

type mappings struct {
	general         *GeneralConfig
	computeResource *map[string]map[string]interface{}
}

// Mapping is the mapping of the terraform resources
var mapping *mappings

// GetMapping returns the mapping of the terraform resources
func getMapping(provider providers.Provider) (*mappings, error) {
	if mapping != nil {
		return mapping, nil
	}
	err := loadMapping(provider)
	if err != nil {
		return nil, err
	}
	return mapping, nil
}

func loadMapping(provider providers.Provider) error {
	if mapping != nil {
		return nil
	}
	mappingFolder := fmt.Sprintf("internal/terraform/%s/mappings", provider)
	files, err := os.ReadDir(mappingFolder)
	if err != nil {
		return err
	}

	mergedMappings := &MappingsFromFiles{
		General:         make(map[string]interface{}),
		ComputeResource: make(map[string]interface{}),
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		yamlFile, err := os.ReadFile(filepath.Join(mappingFolder, file.Name()))
		if err != nil {
			return err
		}
		var currentMapping map[string]interface{}
		err = yaml.Unmarshal(yamlFile, &currentMapping)
		if err != nil {
			return err
		}

		if err != nil {
			return err
		}

		generalI := currentMapping["general"]
		if generalI != nil {
			generalMapping, ok := generalI.(map[string]interface{})
			if !ok {
				return fmt.Errorf("general mapping is not a map[string]interface{}")
			}
			for k, v := range generalMapping {
				mergedMappings.General[k] = v
			}

		}

		computeMappingI := currentMapping["compute_resource"]
		if computeMappingI != nil {
			computeResourceMapping, ok := computeMappingI.(map[string]interface{})
			if !ok {
				return fmt.Errorf("compute_resource mapping is not a map[string]interface{}")
			}
			for k, v := range computeResourceMapping {
				mergedMappings.ComputeResource[k] = v
			}
		}
	}

	generalConfig, err := convertToGeneralConfig(mergedMappings.General)
	if err != nil {
		return err
	}

	computeResourceMapping, err := convertMapToMapOfMaps(mergedMappings.ComputeResource)
	if err != nil {
		return err
	}

	mapping = &mappings{
		general:         generalConfig,
		computeResource: &computeResourceMapping,
	}

	return nil
}

func getMappingProperties(mapping map[string]interface{}) map[string]interface{} {
	propertiesI, ok := mapping["properties"]
	if !ok {
		properties, err := convertInterfaceToMap(mapping)
		if err != nil {
			panic(err)
		}
		return properties
	}
	properties, err := convertInterfaceToMap(propertiesI)
	if err != nil {
		panic(err)
	}
	return properties
}
