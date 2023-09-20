package plan

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
	SizeGb           decimal.Decimal
	IsSSD            bool
	OverridePriority int
	Key              string
}

func applyReference(valueFound string, propertyMapping *PropertyDefinition, context *tfContext) (interface{}, error) {
	if propertyMapping == nil || propertyMapping.Reference == nil {
		return valueFound, nil
	}
	reference := propertyMapping.Reference
	valueTransformed, err := resolveReference(valueFound, reference, context)
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
			valueArray, err := utils.GetJSON(reference.Property, item)
			if err != nil {
				return nil, errors.Wrapf(err, "Cannot find property %v in file %v", reference.Property, reference.JSONFile)
			}
			if len(valueArray) != 0 {
				if valueArray[0] == nil {
					return nil, fmt.Errorf("Cannot find property %v in file %v", reference.Property, reference.JSONFile)
				}
				value = valueArray[0]
			} else {
				return nil, fmt.Errorf("Cannot find property %v in file %v", reference.Property, reference.JSONFile)
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
		defaultDiskType := generalMappings.DiskTypes.Default
		if defaultDiskType != nil {
			return defaultDiskType, nil
		}
		return SSD, nil
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
			referencedItems, err := getJSON(path, *TfPlan)
			if err != nil {
				errW := errors.Wrapf(err, "Cannot find referenced path in terraform plan: '%v'", path)
				return nil, errW
			}
			for _, referencedItem := range referencedItems {
				if reference.Property != "" {
					value, err := utils.GetJSON(reference.Property, referencedItem)
					if err != nil {
						return nil, errors.Wrapf(err, "Cannot find property %v in path %v", reference.Property, path)
					}
					if len(value) != 0 {
						if value[0] != nil {
							return value[0], nil
						}
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
	if reference.ReturnPath {
		return key, nil
	}
	return key, nil
}

func applyRegex(valueFound string, propertyMapping *PropertyDefinition, context *tfContext) (interface{}, error) {
	if propertyMapping == nil || propertyMapping.Regex == nil {
		return valueFound, nil
	}
	regex := *propertyMapping.Regex
	valueTransformed, err := resolveRegex(valueFound, regex)
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

func applyValidator(valueFound interface{}, propertyMapping *PropertyDefinition, context *tfContext) error {
	if propertyMapping == nil || propertyMapping.Validator == nil {
		return nil
	}
	validator := propertyMapping.Validator
	err := resolveValidator(valueFound, validator, context)
	return err
}

func resolveValidator(value interface{}, validator *string, context *tfContext) error {
	_, err := getJSON(*validator, value)
	return errors.Wrapf(err, "Cannot validate '%v' value of %v", value, context.ResourceAddress)
}
