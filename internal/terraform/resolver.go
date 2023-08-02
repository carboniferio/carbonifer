package terraform

import (
	"encoding/json"
	"errors"
	"regexp"

	"github.com/PaesslerAG/jsonpath"
	"github.com/carboniferio/carbonifer/internal/data"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type Regex struct {
	Pattern string
	Group   int
}

type Reference struct {
	JsonFile string      `json:"json_file"`
	Property string      `json:"property"`
	General  string      `json:"general"`
	Paths    interface{} `json:"path"`
}

type Storage struct {
	SizeGb decimal.Decimal
	IsSSD  bool
}

func ApplyReference(valueFound interface{}, propertyMapping map[string]interface{}, resourceAdress string) (interface{}, error) {
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
	if property, ok := refMap["paths"]; ok {
		reference.Paths = property.(string)
	}
	if jsonFile, ok := refMap["general"]; ok {
		reference.General = jsonFile.(string)
	}
	valueTransformed, err := ResolveReference(valueFound.(string), reference, resourceAdress)
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
			log.Fatalf("Cannot find key %v in file %v", key, reference.JsonFile)
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
		paths, err := ReadPaths(resourceAddress, reference.Paths, &templatePlaceholders)
		if err != nil {
			return nil, err
		}
		for _, pathI := range paths {
			path, ok := pathI.(string)
			if !ok {
				return nil, errors.New("Cannot convert path to string")
			}
			referencedItem, err := jsonpath.Get(path, *TfPlan)
			if err != nil {
				return nil, err
			}
			if referencedItem != nil {
				if reference.Property != "" {
					referencedItems := referencedItem.([]interface{})
					for _, referencedItem := range referencedItems {
						value := referencedItem.(map[string]interface{})[reference.Property]
						return value, nil
					}
				} else {
					return referencedItem, nil
				}
			}
		}
		return nil, nil
	}
	return key, nil
}

func ApplyRegex(valueFound interface{}, propertyMapping map[string]interface{}, resourceAdressw string) (interface{}, error) {
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
	valueTransformed := ResolveRegex(valueFound.(string), regex)
	return valueTransformed, nil
}

func ResolveRegex(value string, regex Regex) string {
	r, _ := regexp.Compile(regex.Pattern)

	matches := r.FindStringSubmatch(value)

	if len(matches) > 1 {
		return matches[regex.Group]
	} else {
		log.Fatalf("No match found for regex %v in value %v", regex.Pattern, value)
	}
	return ""
}
