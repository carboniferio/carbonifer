package terraform

import (
	"fmt"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

func GetString(key string, resourceAdress string, resource map[string]interface{}, mapping map[string]interface{}) (*string, error) {
	value, err := GetValue(key, resourceAdress, resource, mapping)
	if err != nil {
		return nil, err
	}

	if value == nil {
		log.Debugf("No value found for key %v of resource type %v", key, resourceAdress)
		return nil, nil
	}
	stringValue, ok := value.Value.(string)
	if !ok {
		return nil, fmt.Errorf("Cannot convert value to string: %v", value.Value)
	}
	return &stringValue, nil
}

func GetSlice(key string, resourceType string, resource map[string]interface{}, mapping map[string]interface{}) ([]interface{}, error) {
	results := []interface{}{}
	sliceMappingI := GetMappingProperties(mapping)[key]
	if sliceMappingI == nil {
		return nil, nil
	}
	sliceMapping, ok := sliceMappingI.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Cannot get mapping for %v of resource type %v", key, resourceType)
	}

	// if type exists in mapping and is "list"
	t, ok := sliceMapping["type"]
	if ok && t == "list" {
		// get items property of mapping
		items, ok := sliceMapping["item"]
		if !ok {
			return nil, fmt.Errorf("Cannot get items property of mapping for %v of resource type %v", key, resourceType)
		}
		itemsList, ok := items.([]interface{})
		if !ok {
			return nil, fmt.Errorf("Items is not a list for %v of resource type %v", key, resourceType)
		}
		for _, item := range itemsList {
			result, err := GetStorage(item, resource)
			if err != nil {
				return nil, err
			}

			for _, r := range result {
				results = append(results, r)
			}
		}
	}
	return results, nil
}

func GetStorage(item interface{}, resource map[string]interface{}) ([]Storage, error) {
	itemMap := item.(map[string]interface{})
	pathsProperty := itemMap["paths"]
	paths, err := ReadPaths("storage", pathsProperty)
	if err != nil {
		return nil, fmt.Errorf("Cannot get paths for storage: %v", err)
	}
	for _, path := range paths {
		if strings.Contains(path.(string), "${") {
			// TODO ${template_config}
			continue
		}
		storageRaw, err := jsonpath.Get(path.(string), resource)
		if err != nil {
			return nil, fmt.Errorf("Cannot get storage: %v", err)
		}
		if storageRaw == nil {
			continue
		}
		storageI, ok := storageRaw.([]interface{})
		if !ok {
			storageI = []interface{}{storageRaw}
		}
		if len(storageI) == 0 {
			continue
		}
		storages, err := ConvertInterfaceSlicesToMapSlice(storageI)
		if err != nil {
			return nil, fmt.Errorf("Cannot convert storage: %v", err)
		}
		storagesResults := []Storage{}
		for _, storageMap := range storages {
			storage, err := getStorage(storageMap, itemMap)
			if err != nil {
				return nil, err
			}

			if storage != nil {
				storagesResults = append(storagesResults, *storage)
			}
		}
		return storagesResults, nil

	}
	return nil, nil
}

type ValueWithUnit struct {
	Value interface{}
	Unit  *string
}

func getStorage(storageMap map[string]interface{}, itemMap map[string]interface{}) (*Storage, error) {
	storageSize, err := GetValue("size", "storage", storageMap, itemMap)
	if err != nil {
		return nil, err
	}

	if storageSize == nil {
		return nil, nil
	}
	storageSizeGb, err := decimal.NewFromString(fmt.Sprintf("%v", storageSize.Value))
	if err != nil {
		return nil, fmt.Errorf("Cannot convert storage size to int: %v", storageSize)
	}
	storageType, err := GetValue("type", "storage", storageMap, itemMap)
	if err != nil {
		return nil, err
	}
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

	return &Storage{
		SizeGb: storageSizeGb,
		IsSSD:  isSSD,
	}, nil
}

type KeyValue struct {
	Key   string
	Value interface{}
}

func ReadPaths(resourceAdress string, pathsProperty interface{}, pathTemplateValuesParams ...*map[string]string) ([]interface{}, error) {
	var paths []interface{}
	if pathsStr, ok := pathsProperty.(string); ok {
		paths = []interface{}{pathsStr}
	} else if pathsSlice, ok := pathsProperty.([]interface{}); ok {
		paths = pathsSlice
	} else {
		return nil, fmt.Errorf("paths is neither a string nor a slice of strings for %v", resourceAdress)
	}

	for _, pathTemplateValues := range pathTemplateValuesParams {
		for path := range paths {
			pathStr := paths[path].(string)
			for key, value := range *pathTemplateValues {
				pathStr = strings.ReplaceAll(pathStr, "${"+key+"}", value)
			}
			paths[path] = pathStr
		}
	}
	return paths, nil
}

func GetValue(key string, resourceAdress string, resource map[string]interface{}, mapping map[string]interface{}) (*ValueWithUnit, error) {

	propertyMappings := GetPropertyMappings(mapping, key, resourceAdress)
	if propertyMappings == nil {
		log.Debugf("No property mapping found for key %v of resource type %v", key, resourceAdress)
		return nil, nil
	}

	var valueFound interface{}
	for _, propertyMapping := range propertyMappings {
		pathProperty := propertyMapping["paths"]
		paths, err := ReadPaths(resourceAdress, pathProperty)
		if err != nil {
			return nil, err
		}
		var unit *string
		unitI := propertyMapping["unit"]
		if unitI != nil {
			unitStr, ok := unitI.(string)
			if !ok {
				return nil, fmt.Errorf("Cannot convert unit to string: %v", unitI)
			}
			unit = &unitStr
		}

		for _, path := range paths {
			if valueFound != nil {
				break
			}
			if strings.Contains(path.(string), "${") {
				// TODO ${template_config}
				log.Warnf("Template config not yet implemented in %v", path.(string))
				continue
			}
			valueFounds, err := jsonpath.Get(path.(string), resource)
			if err != nil {
				return nil, err
			}
			if valueFounds != nil {
				valueFoundSlice, ok := valueFounds.([]interface{})
				if ok {
					if len(valueFoundSlice) > 1 {
						return nil, fmt.Errorf("Found more than one value for property %v of resource type %v", key, resourceAdress)
					}
					if len(valueFoundSlice) == 0 {
						return nil, fmt.Errorf("No value found for property %v of resource type %v", key, resourceAdress)
					}
					valueFound = valueFoundSlice[0]
				} else {
					valueFound = valueFounds
				}
			}
		}

		if valueFound != nil {
			valueFound, err = ApplyRegex(valueFound, propertyMapping, resourceAdress)
			if err != nil {
				return nil, err
			}
			valueFound, err = ApplyReference(valueFound, propertyMapping, resourceAdress)
			if err != nil {
				return nil, err
			}
		}

		if valueFound != nil {
			return &ValueWithUnit{
				Value: valueFound,
				Unit:  unit,
			}, nil
		}
	}

	if valueFound == nil {
		defaultValue, err := GetDefaultValue(key, resourceAdress, resource, mapping)
		if err != nil {
			return nil, err
		}

		if defaultValue != nil {
			return defaultValue, nil
		}
	}

	return nil, nil
}

func GetDefaultValue(key string, resourceAdress string, resource map[string]interface{}, mapping map[string]interface{}) (*ValueWithUnit, error) {
	propertyMappings := GetPropertyMappings(mapping, key, resourceAdress)
	if propertyMappings == nil {
		log.Debugf("No property mapping found for key %v of resource type %v", key, resourceAdress)
		return nil, nil
	}

	var valueFound interface{}
	var unit *string
	for _, propertyMapping := range propertyMappings {
		if valueFound != nil {
			break
		}
		defaultValue, ok := propertyMapping["default"]
		if !ok {
			continue
		}
		valueFound = defaultValue
		unitI := propertyMapping["unit"]
		if unitI != nil {
			unitStr, ok := unitI.(string)
			if !ok {
				return nil, fmt.Errorf("Cannot convert unit to string: %v", unitI)
			}
			unit = &unitStr
		}
	}
	if valueFound != nil {
		return &ValueWithUnit{
			Value: valueFound,
			Unit:  unit,
		}, nil
	}
	return nil, nil

}
