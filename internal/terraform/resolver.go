package terraform

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/carboniferio/carbonifer/internal/data"
	"github.com/carboniferio/carbonifer/internal/utils"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

type storage struct {
	SizeGb decimal.Decimal
	IsSSD  bool
}

func applyReference(valueFound interface{}, propertyMapping *PropertyDefinition, context *tfContext) (interface{}, error) {
	if propertyMapping == nil || propertyMapping.Reference == nil {
		return valueFound, nil
	}

	reference := propertyMapping.Reference
	valueTransformed, err := resolveReference(valueFound.(string), reference, context)
	return valueTransformed, err
}

func resolveReference(key string, reference *Reference, context *tfContext) (interface{}, error) {
	generalMappings := (*globalMappings.General)[context.Provider]
	if reference.JSONFile != "" {
		filename, ok := (*generalMappings.JSONData)[reference.JSONFile]
		if !ok {
			log.Fatalf("Cannot find file %v in general.json_data", reference.JSONFile)
		}
		byteValue := data.ReadDataFile(filename.(string))
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
		for providerDiskType, diskType := range *generalMappings.DiskTypes.Types {
			if providerDiskType == key {
				return diskType, nil
			}
		}
		if generalMappings.DiskTypes.Default != nil {
			return generalMappings.DiskTypes.Default, nil
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
				errW := errors.Wrapf(err, "Cannot find referenced path in terraform plan: '%v'", path)
				return nil, errW
			}
			for _, referencedItem := range referencedItems {
				if reference.Property != "" {
					value := referencedItem.(map[string]interface{})[reference.Property]
					if value != nil {
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

func applyRegex(valueFound interface{}, propertyMapping *PropertyDefinition, context *tfContext) (interface{}, error) {
	if propertyMapping == nil || propertyMapping.Regex == nil {
		return valueFound, nil
	}
	regex := *propertyMapping.Regex
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
