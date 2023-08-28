package terraform

import (
	"fmt"

	"github.com/pkg/errors"
)

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

func convertInterfaceToMap(input interface{}) (map[string]interface{}, error) {
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

func convertInterfaceSlicesToMapSlice(input []interface{}) ([]map[string]interface{}, error) {
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

func convertMapToMapOfMaps(in map[string]interface{}) (map[string]map[string]interface{}, error) {
	out := make(map[string]map[string]interface{})
	for key, value := range in {
		strValue, ok := value.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("cannot convert map value of type %T to map[string]interface{}", value)
		}
		out[key] = strValue
	}
	return out, nil
}
