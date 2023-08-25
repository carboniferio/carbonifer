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

type Storage struct {
	SizeGb decimal.Decimal
	IsSSD  bool
}

func ApplyReference(valueFound interface{}, propertyMapping map[string]interface{}, resourceAddress string) (interface{}, error) {
	refI, ok := propertyMapping["reference"]
	if !ok {
		return valueFound, nil
	}
	refMap, err := ConvertInterfaceToMap(refI)
	if err != nil {
		return nil, err
	}
	reference := Reference{}
	if jsonFile, ok := refMap["json_file"]; ok {
		reference.JsonFile = jsonFile.(string)
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
	valueTransformed, err := ResolveReference(valueFound.(string), reference, resourceAddress)
	return valueTransformed, err
}

func ResolveReference(key string, reference Reference, resourceAddress string) (interface{}, error) {
	if reference.JsonFile != "" {
		filename := GeneralMappingConfig.JSONData[reference.JsonFile]
		byteValue := data.ReadDataFile(filename)
		var fileMap map[string]interface{}
		err := json.Unmarshal([]byte(byteValue), &fileMap)
		if err != nil {
			log.Fatal(err)
		}
		item, ok := fileMap[key]
		if !ok {
			// Not an error, for example gcp compute type can be a regex
			log.Debugf("Cannot find key %v in file %v", key, reference.JsonFile)
			return nil, nil
		}
		var value interface{}
		property := reference.Property
		if property != "" {
			value, ok = item.(map[string]interface{})[reference.Property]
			if !ok {
				log.Fatalf("Cannot find property %v in file %v", reference.Property, reference.JsonFile)
			}
		}
		return value, nil
	}
	if reference.General != "" {
		for providerDiskType, diskType := range GeneralMappingConfig.DiskTypes.Types {
			if providerDiskType == key {
				return diskType, nil
			}
		}
		if GeneralMappingConfig.DiskTypes.Default != "" {
			return GeneralMappingConfig.DiskTypes.Default, nil
		}
		return "ssd", nil
	}
	if reference.Paths != nil {
		templatePlaceholders := map[string]string{
			"key": key,
		}
		paths, err := ReadPaths(reference.Paths, &templatePlaceholders)
		if err != nil {
			return nil, err
		}
		for _, path := range paths {
			referencedItems, err := utils.JsonGet(path, *TfPlan)
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

func ApplyRegex(valueFound interface{}, propertyMapping map[string]interface{}, resourceAddressw string) (interface{}, error) {
	regexI, ok := propertyMapping["regex"]
	if !ok {
		return valueFound, nil
	}
	regexMap, err := ConvertInterfaceToMap(regexI)
	if err != nil {
		return nil, err
	}

	regex := Regex{
		Pattern: regexMap["pattern"].(string),
		Group:   int(regexMap["group"].(float64)),
	}
	valueTransformed, err := ResolveRegex(valueFound.(string), regex)
	return valueTransformed, err
}

func ResolveRegex(value string, regex Regex) (string, error) {
	r, _ := regexp.Compile(regex.Pattern)

	matches := r.FindStringSubmatch(value)

	if len(matches) > 1 {
		return matches[regex.Group], nil
	} else {
		return "", fmt.Errorf("No match found for regex %v in value %v", regex.Pattern, value)
	}
}
