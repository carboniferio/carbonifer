package terraform

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/carboniferio/carbonifer/internal/data"
	"github.com/carboniferio/carbonifer/internal/utils"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type storage struct {
	SizeGb decimal.Decimal
	IsSSD  bool
}

func applyReference(valueFound interface{}, propertyMapping map[string]interface{}, resourceAddress string) (interface{}, error) {
	refI, ok := propertyMapping["reference"]
	if !ok {
		return valueFound, nil
	}
	refMap, err := convertInterfaceToMap(refI)
	if err != nil {
		return nil, err
	}
	reference := Reference{}
	if jsonFile, ok := refMap["json_file"]; ok {
		reference.JSONFile = jsonFile.(string)
	}
	if property, ok := refMap["property"]; ok {
		reference.Property = property.(string)
	}
	if returnPath, ok := refMap["return_path"]; ok {
		reference.ReturnPath = returnPath.(bool)
	}
	if propertyI, ok := refMap["paths"]; ok {
		switch property := propertyI.(type) {
		case []string:
			reference.Paths = property
		case []interface{}:
			reference.Paths = utils.ConvertInterfaceListToStringList(property)
		case string:
			reference.Paths = []string{property}
		default:
			return nil, errors.New("Cannot convert paths to string or []string")
		}
	}
	if jsonFile, ok := refMap["general"]; ok {
		reference.General = jsonFile.(string)
	}
	valueTransformed, err := resolveReference(valueFound.(string), reference, resourceAddress)
	return valueTransformed, err
}

func resolveReference(key string, reference Reference, resourceAddress string) (interface{}, error) {
	if reference.JSONFile != "" {
		filename := mapping.general.JSONData[reference.JSONFile]
		byteValue := data.ReadDataFile(filename)
		var fileMap map[string]interface{}
		err := json.Unmarshal([]byte(byteValue), &fileMap)
		if err != nil {
			log.Fatal(err)
		}
		item, ok := fileMap[key]
		if !ok {
			// Not an error, for example gcp compute type can be a regex
			log.Debugf("Cannot find key %v in file %v", key, reference.JSONFile)
			return nil, nil
		}
		var value interface{}
		property := reference.Property
		if property != "" {
			value, ok = item.(map[string]interface{})[reference.Property]
			if !ok {
				log.Fatalf("Cannot find property %v in file %v", reference.Property, reference.JSONFile)
			}
		}
		return value, nil
	}
	if reference.General != "" {
		for providerDiskType, diskType := range mapping.general.DiskTypes.Types {
			if providerDiskType == key {
				return diskType, nil
			}
		}
		if mapping.general.DiskTypes.Default != "" {
			return mapping.general.DiskTypes.Default, nil
		}
		return "ssd", nil
	}
	if reference.Paths != nil {
		templatePlaceholders := map[string]string{
			"key": key,
		}
		paths, err := readPaths(reference.Paths, &templatePlaceholders)
		if err != nil {
			return nil, err
		}
		for _, path := range paths {
			referencedItems, err := utils.GetJSON(path, *TfPlan)
			if err != nil {
				return nil, err
			}
			if referencedItems != nil {
				if reference.Property != "" {
					for _, referencedItem := range referencedItems {
						value := referencedItem.(map[string]interface{})[reference.Property]
						return value, nil
					}
				} else if reference.ReturnPath {
					return path, nil
				} else {
					return referencedItems, nil
				}
			}
		}
		return nil, nil
	}
	return key, nil
}

func applyRegex(valueFound interface{}, propertyMapping map[string]interface{}, resourceAddressw string) (interface{}, error) {
	regexI, ok := propertyMapping["regex"]
	if !ok {
		return valueFound, nil
	}
	regexMap, err := convertInterfaceToMap(regexI)
	if err != nil {
		return nil, err
	}

	regex := Regex{
		Pattern: regexMap["pattern"].(string),
		Group:   int(regexMap["group"].(float64)),
	}
	valueTransformed, err := resolveRegex(valueFound.(string), regex)
	return valueTransformed, err
}

func resolveRegex(value string, regex Regex) (string, error) {
	r, _ := regexp.Compile(regex.Pattern)

	matches := r.FindStringSubmatch(value)

	if len(matches) > 1 {
		return matches[regex.Group], nil
	}
	return "", fmt.Errorf("No match found for regex %v in value %v", regex.Pattern, value)

}
