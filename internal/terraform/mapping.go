package terraform

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

type Mappings struct {
	General         map[string]interface{} `yaml:"general"`
	ComputeResource map[string]interface{} `yaml:"compute_resource"`
}

func convertInnerMaps(m map[string]interface{}) (map[string]interface{}, error) {
	newMap := make(map[string]interface{})
	for k, v := range m {
		if innerMap, ok := v.(map[interface{}]interface{}); ok {
			newInnerMap := make(map[string]interface{})
			for innerK, innerV := range innerMap {
				if innerKStr, ok := innerK.(string); ok {
					newInnerMap[innerKStr] = innerV
				} else {
					return nil, fmt.Errorf("invalid key type: expected string, got %T", innerK)
				}
			}
			newMap[k] = newInnerMap
		} else {
			newMap[k] = v
		}
	}
	return newMap, nil
}

func LoadMapping(mappingFolder string) (*Mappings, error) {
	files, err := os.ReadDir(mappingFolder)
	if err != nil {
		return nil, err
	}

	mergedMappings := &Mappings{
		General:         make(map[string]interface{}),
		ComputeResource: make(map[string]interface{}),
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		yamlFile, err := os.ReadFile(filepath.Join(mappingFolder, file.Name()))
		if err != nil {
			return nil, err
		}
		var currentMapping map[string]interface{}
		err = yaml.Unmarshal(yamlFile, &currentMapping)
		if err != nil {
			return nil, err
		}

		if err != nil {
			return nil, err
		}

		generalI := currentMapping["general"]
		if generalI != nil {
			generalMapping, ok := generalI.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("general mapping is not a map[string]interface{}")
			}
			for k, v := range generalMapping {
				mergedMappings.General[k] = v
			}

		}

		computeMappingI := currentMapping["compute_resource"]
		if computeMappingI != nil {
			computeResourceMapping, ok := computeMappingI.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("compute_resource mapping is not a map[string]interface{}")
			}
			for k, v := range computeResourceMapping {
				mergedMappings.ComputeResource[k] = v
			}
		}
	}

	return mergedMappings, nil
}

func GetMappingProperties(mapping map[string]interface{}) map[string]interface{} {
	propertiesI, ok := mapping["properties"]
	if !ok {
		properties, err := ConvertInterfaceToMap(mapping)
		if err != nil {
			panic(err)
		}
		return properties
	}
	properties, err := ConvertInterfaceToMap(propertiesI)
	if err != nil {
		panic(err)
	}
	return properties
}

func ConvertInterfaceToMap(input interface{}) (map[string]interface{}, error) {
	switch typedInput := input.(type) {
	case map[string]interface{}:
		return typedInput, nil
	case map[interface{}]interface{}:
		strKeysMap, err := convertMapKeysToStrings(typedInput)
		if err != nil {
			return nil, err
		}
		return convertInnerMaps(strKeysMap)
	default:
		return nil, fmt.Errorf("input is neither map[string]interface{} nor map[interface{}]interface{}")
	}
}

func ConvertInterfaceSlicesToMapSlice(input []interface{}) ([]map[string]interface{}, error) {
	var output []map[string]interface{}
	for _, element := range input {
		switch typedElement := element.(type) {
		case map[string]interface{}:
			output = append(output, typedElement)
		case map[interface{}]interface{}:
			strKeysMap, err := convertMapKeysToStrings(typedElement)
			if err != nil {
				return nil, err
			}
			convertedMap, err := convertInnerMaps(strKeysMap)
			if err != nil {
				return nil, err
			}
			output = append(output, convertedMap)
		default:
			return nil, errors.Errorf("input is neither map[string]interface{} nor map[interface{}]interface{} : %T", element)
		}
	}
	return output, nil
}

func convertMapKeysToStrings(in map[interface{}]interface{}) (map[string]interface{}, error) {
	out := make(map[string]interface{})
	for key, value := range in {
		strKey, ok := key.(string)
		if !ok {
			return nil, fmt.Errorf("cannot convert map key of type %T to string", key)
		}
		out[strKey] = value
	}
	return out, nil
}
