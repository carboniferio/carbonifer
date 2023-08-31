package terraform

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// tfContext is the context of a terraform resource
type tfContext struct {
	Resource        map[string]interface{} // Json of the terraform plan resource
	Mapping         *ResourceMapping       // Mapping of the resource type
	ResourceAddress string                 // Address of the resource in tf plan
	ParentContext   *tfContext             // Parent context
	Provider        providers.Provider
}

func getString(key string, context *tfContext) (*string, error) {
	value, err := getValue(key, context)
	if err != nil {
		return nil, err
	}

	if value == nil {
		log.Debugf("No value found for key %v of resource type %v", key, context.ResourceAddress)
		return nil, nil
	}
	stringValue, ok := value.Value.(string)
	if !ok {
		return nil, fmt.Errorf("Cannot convert value to string: %v : %T", value.Value, value.Value)
	}
	return &stringValue, nil
}

func getSlice(key string, context *tfContext) ([]interface{}, error) {
	results := []interface{}{}

	sliceMappings := (*context.Mapping.Properties)[key]

	// Check we are well working on a list
	for _, sliceMapping := range sliceMappings {

		if sliceMapping.ValueType != nil && *sliceMapping.ValueType != "list" {
			return nil, fmt.Errorf("Cannot get slice for %v if resource '.type' is not 'list'", key)
		}

		// get mapping of items of the list
		mappingItems := sliceMapping.Item
		if mappingItems == nil {
			return nil, fmt.Errorf("Items is not a list for %v of resource type %v", key, context.ResourceAddress)
		}
		for _, itemMapping := range *mappingItems {
			context := tfContext{
				Resource:        context.Resource,
				Mapping:         &itemMapping,
				ResourceAddress: context.ResourceAddress + "." + key,
				ParentContext:   context,
				Provider:        context.Provider,
			}
			itemResults, err := getSliceItems(context)
			if err != nil {
				return nil, err
			}
			results = append(results, itemResults...)
		}
	}

	return results, nil
}

func getSliceItems(context tfContext) ([]interface{}, error) {
	itemMapping := context.Mapping
	results := []interface{}{}
	paths, err := readPaths(itemMapping.Paths)
	if err != nil {
		return nil, fmt.Errorf("Cannot get paths for %v: %v", context.ResourceAddress, err)
	}

	for _, pathRaw := range paths {
		path := pathRaw
		if strings.Contains(pathRaw, "${") {
			path, err = resolvePlaceholders(path, context.ParentContext)
			if err != nil {
				return nil, err
			}
		}
		jsonResults, err := utils.GetJSON(path, context.Resource)
		if err != nil {
			return nil, errors.Wrapf(err, "Cannot get item: %v", path)
		}
		// if no result, try to get it from the whole plan
		if len(jsonResults) == 0 && TfPlan != nil {
			jsonResults, err = utils.GetJSON(path, *TfPlan)
			if err != nil {
				return nil, errors.Wrapf(err, "Cannot get item: %v", path)
			}
		}
		for _, jsonResultsI := range jsonResults {
			switch jsonResults := jsonResultsI.(type) {
			case map[string]interface{}:
				result, err := getItem(context, itemMapping, jsonResults)
				if err != nil {
					return nil, err
				}
				results = append(results, result)
			case []interface{}:
				for _, jsonResultI := range jsonResults {
					jsonResultI, ok := jsonResultI.(map[string]interface{})
					if !ok {
						return nil, errors.Errorf("Cannot convert jsonResultI to map[string]interface{}: %v", jsonResultI)
					}
					result, err := getItem(context, itemMapping, jsonResultI)
					if err != nil {
						return nil, err
					}
					results = append(results, result)
				}
			default:
				return nil, errors.Errorf("Not an map or an array of maps: %T", jsonResultsI)
			}
		}
	}
	return results, nil
}

func getItem(context tfContext, itemMappingProperties *ResourceMapping, jsonResultI map[string]interface{}) (interface{}, error) {
	result := map[string]interface{}{}
	for key := range *itemMappingProperties.Properties {
		if key == "paths" {
			continue
		}
		itemContext := tfContext{
			Resource:        jsonResultI,
			Mapping:         itemMappingProperties,
			ResourceAddress: context.ResourceAddress,
			ParentContext:   &context,
			Provider:        context.Provider,
		}
		property, err := getValue(key, &itemContext)
		if err != nil {
			return nil, err
		}
		result[key] = property
	}
	return result, nil
}

type valueWithUnit struct {
	Value interface{}
	Unit  *string
}

func readPaths(pathsProperty interface{}, pathTemplateValuesParams ...*map[string]string) ([]string, error) {
	paths := []string{}
	if pathsProperty == nil {
		return paths, nil
	}

	switch pathTyped := pathsProperty.(type) {
	case string:
		paths = []string{pathTyped}
	case []string:
		paths = append(paths, pathTyped...)
	case []interface{}:
		for _, pathI := range pathTyped {
			pathStr, ok := pathI.(string)
			if !ok {
				return nil, errors.Errorf("Cannot convert path to string: %T", pathI)
			}
			paths = append(paths, pathStr)
		}
	default:
		return nil, errors.Errorf("Cannot convert paths to string or []string: %T", pathsProperty)
	}

	for _, pathTemplateValues := range pathTemplateValuesParams {
		for i, path := range paths {
			pathStr := path
			for key, value := range *pathTemplateValues {
				pathStr = strings.ReplaceAll(pathStr, "${"+key+"}", value)
			}
			paths[i] = pathStr
		}
	}
	return paths, nil
}

func getValue(key string, context *tfContext) (*valueWithUnit, error) {

	var valueFound interface{}
	propertiesMappings := (*context.Mapping.Properties)[key]
	for _, propertyMapping := range propertiesMappings {
		paths, err := readPaths(propertyMapping.Paths)
		if err != nil {
			return nil, err
		}
		unit := propertyMapping.Unit

		for _, pathRaw := range paths {
			if valueFound != nil {
				break
			}
			path := pathRaw
			if strings.Contains(pathRaw, "${") {
				path, err = resolvePlaceholders(path, context)
				if err != nil {
					return nil, err
				}
			}
			valueFounds, err := utils.GetJSON(path, context.Resource)
			if err != nil {
				return nil, err
			}
			if len(valueFounds) == 0 && TfPlan != nil {
				// Try to resolve it against the whole plan
				valueFounds, err = utils.GetJSON(path, *TfPlan)
				if err != nil {
					return nil, err
				}
			}
			if len(valueFounds) > 0 {
				if len(valueFounds) > 1 {
					return nil, fmt.Errorf("Found more than one value for property %v of resource type %v", key, context.ResourceAddress)
				}
				if valueFounds[0] == nil {
					continue
				}
				valueFound = valueFounds[0]
			}
		}

		if valueFound != nil {
			valueFoundStr, ok := valueFound.(string)
			if ok {
				valueFound, err = applyRegex(valueFoundStr, &propertyMapping, context)
				if err != nil {
					return nil, err
				}
			}
			valueFoundStr, ok = valueFound.(string)
			if ok {
				valueFound, err = applyReference(valueFoundStr, &propertyMapping, context)
				if err != nil {
					return nil, err
				}
			}
		}

		// if value is an expression (map[string]interface{}), resolve it
		valueFoundMap, ok := valueFound.(map[string]interface{})
		if ok {
			valueFound, err = getValueOfExpression(valueFoundMap, context)
			if err != nil {
				return nil, err
			}
		}

		if valueFound != nil {
			return &valueWithUnit{
				Value: valueFound,
				Unit:  unit,
			}, nil
		}
	}

	if valueFound == nil {
		defaultValue, err := getDefaultValue(key, context)
		if err != nil {
			return nil, err
		}

		if defaultValue != nil {
			return defaultValue, nil
		}
	}

	return nil, nil
}

func resolvePlaceholders(input string, context *tfContext) (string, error) {
	placeholderPattern := `\${([^}]+)}`

	// Compile the regular expression
	rx := regexp.MustCompile(placeholderPattern)

	// Find all matches in the input string
	matches := rx.FindAllStringSubmatch(input, -1)

	// Create a map to store resolved expressions
	resolvedExpressions := make(map[string]string)

	// Iterate through the matches and resolve expressions
	for _, match := range matches {
		placeholder := match[0]
		expression := match[1]
		resolved, err := resolvePlaceholder(expression, context)
		if err != nil {
			return input, err
		}
		resolvedExpressions[placeholder] = resolved
	}

	// Replace placeholders in the input string with resolved expressions
	replacerStrings := make([]string, 0, len(resolvedExpressions)*2)
	for placeholder, resolved := range resolvedExpressions {
		replacerStrings = append(replacerStrings, placeholder, resolved)
	}

	replacer := strings.NewReplacer(replacerStrings...)
	resolvedString := replacer.Replace(input)
	return resolvedString, nil
}

func resolvePlaceholder(expression string, context *tfContext) (string, error) {
	result := ""
	if strings.HasPrefix(expression, "this.") {
		thisProperty := strings.TrimPrefix(expression, "this")
		resource := context.Resource
		value, err := utils.GetJSON(thisProperty, resource)
		if err != nil {
			return "", errors.Wrapf(err, "Cannot get value for variable %s", expression)
		}
		if value == nil {
			return "", errors.Errorf("No value found for variable %s", expression)
		}
		return fmt.Sprintf("%v", value[0]), err
	} else if strings.HasPrefix(expression, "config.") {
		configProperty := strings.TrimPrefix(expression, "config.")
		value := viper.GetFloat64(configProperty)
		return fmt.Sprintf("%v", value), nil
	}
	variable, err := getVariable(expression, context)
	if err != nil {
		return "", err
	}
	if variable != nil {
		result = fmt.Sprintf("%v", variable)
	}
	return result, nil
}

func getDefaultValue(key string, context *tfContext) (*valueWithUnit, error) {
	propertyMappings, ok := (*context.Mapping.Properties)[key]
	if !ok {
		log.Debugf("No property mapping found for key %v of resource type %v", key, context.ResourceAddress)
		return nil, nil
	}

	for _, propertyMapping := range propertyMappings {
		if propertyMapping.Default != nil {

			valueFound := propertyMapping.Default
			unit := propertyMapping.Unit
			var err error
			valueFoundStr, ok := valueFound.(string)
			if ok {
				valueFound, err = applyRegex(valueFoundStr, &propertyMapping, context)
				if err != nil {
					return nil, err
				}
			}
			valueFoundStr, ok = valueFound.(string)
			if ok {
				valueFound, err = applyReference(valueFoundStr, &propertyMapping, context)
				if err != nil {
					return nil, err
				}
			}

			if valueFound != nil {
				return &valueWithUnit{
					Value: valueFound,
					Unit:  unit,
				}, nil
			}
			return nil, nil
		}
	}
	return nil, nil

}

func getVariable(name string, context *tfContext) (interface{}, error) {
	variablesMappings := context.Mapping.Variables
	if variablesMappings == nil {
		return nil, nil
	}
	variableContext := tfContext{
		Resource:        context.Resource,
		Mapping:         variablesMappings,
		ResourceAddress: context.ResourceAddress + ".variables",
		ParentContext:   context.ParentContext,
		Provider:        context.Provider,
	}
	value, err := getValue(name, &variableContext)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, fmt.Errorf("Cannot get variable : %v", name)
	}
	return value.Value, nil

}
