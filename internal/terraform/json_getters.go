package terraform

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/carboniferio/carbonifer/internal/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Context object
type TFContext struct {
	Resource        map[string]interface{} // Json of the terraform plan resource
	Mapping         map[string]interface{} // Mapping of the resource type
	ResourceAddress string                 // Address of the resource in tf plan
	ParentContext   *TFContext             // Parent context
}

func GetString(key string, context TFContext) (*string, error) {
	value, err := GetValue(key, context)
	if err != nil {
		return nil, err
	}

	if value == nil {
		log.Debugf("No value found for key %v of resource type %v", key, context.ResourceAddress)
		return nil, nil
	}
	stringValue, ok := value.Value.(string)
	if !ok {
		return nil, fmt.Errorf("Cannot convert value to string: %v", value.Value)
	}
	return &stringValue, nil
}

func GetSlice(key string, context TFContext) ([]interface{}, error) {
	results := []interface{}{}

	sliceMappingI := GetMappingProperties(context.Mapping)[key]
	if sliceMappingI == nil {
		return nil, nil
	}
	sliceMapping, ok := sliceMappingI.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Cannot get mapping for %v of resource type %v", key, context.ResourceAddress)
	}

	// Check we are well working on a list
	t, ok := sliceMapping["type"]
	if !ok || t != "list" {
		return nil, fmt.Errorf("Cannot get slice for %v if resource '.type' is not 'list'", key)
	}

	// get mapping of items of the list
	mappingItemsI, ok := sliceMapping["item"]
	if !ok {
		return nil, fmt.Errorf("Cannot get items property of mapping for %v of resource type %v", key, context.ResourceAddress)
	}
	mappingItems, ok := mappingItemsI.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Items is not a list for %v of resource type %v", key, context.ResourceAddress)
	}
	for _, itemMappingI := range mappingItems {
		itemMapping := itemMappingI.(map[string]interface{})
		context := TFContext{
			Resource:        context.Resource,
			Mapping:         itemMapping,
			ResourceAddress: context.ResourceAddress + "." + key,
			ParentContext:   &context,
		}
		itemResults, err := GetSliceItems(context)
		if err != nil {
			return nil, err
		}
		results = append(results, itemResults...)
	}

	return results, nil
}

func GetSliceItems(context TFContext) ([]interface{}, error) {
	itemMapping := context.Mapping
	results := []interface{}{}
	pathsProperty := itemMapping["paths"]
	paths, err := ReadPaths(pathsProperty)
	if err != nil {
		return nil, fmt.Errorf("Cannot get paths for %v: %v", context.ResourceAddress, err)
	}

	itemMappingProperties := GetMappingProperties(itemMapping)

	for _, pathRaw := range paths {
		path := pathRaw
		if strings.Contains(pathRaw, "${") {
			path, err = resolvePlaceholders(path, *context.ParentContext)
			if err != nil {
				return nil, err
			}
		}
		jsonResults, err := utils.JsonGet(path, context.Resource)
		if err != nil {
			return nil, errors.Wrapf(err, "Cannot get item: %v", path)
		}
		// if no result, try to get it from the whole plan
		if len(jsonResults) == 0 {
			jsonResults, err = utils.JsonGet(path, *TfPlan)
			if err != nil {
				return nil, errors.Wrapf(err, "Cannot get item: %v", path)
			}
		}
		for _, jsonResultsI := range jsonResults {
			switch jsonResults := jsonResultsI.(type) {
			case map[string]interface{}:
				result, err := getItem(context, itemMappingProperties, jsonResults)
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
					result, err := getItem(context, itemMappingProperties, jsonResultI)
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

func getItem(context TFContext, itemMappingProperties map[string]interface{}, jsonResultI map[string]interface{}) (interface{}, error) {
	result := map[string]interface{}{}
	for key, _ := range itemMappingProperties {
		if key == "paths" {
			continue
		}
		itemContext := TFContext{
			Resource:        jsonResultI,
			Mapping:         itemMappingProperties,
			ResourceAddress: context.ResourceAddress,
			ParentContext:   &context,
		}
		property, err := GetValue(key, itemContext)
		if err != nil {
			return nil, err
		}
		result[key] = property
	}
	return result, nil
}

type ValueWithUnit struct {
	Value interface{}
	Unit  *string
}

type KeyValue struct {
	Key   string
	Value interface{}
}

func ReadPaths(pathsProperty interface{}, pathTemplateValuesParams ...*map[string]string) ([]string, error) {
	var paths []string
	if pathsProperty == nil {
		return paths, nil
	}

	switch path := pathsProperty.(type) {
	case string:
		paths = []string{path}
	case []string:
		paths = path
	case []interface{}:
		for _, pathI := range path {
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
		for path := range paths {
			pathStr := paths[path]
			for key, value := range *pathTemplateValues {
				pathStr = strings.ReplaceAll(pathStr, "${"+key+"}", value)
			}
			paths[path] = pathStr
		}
	}
	return paths, nil
}

func GetValue(key string, context TFContext) (*ValueWithUnit, error) {

	propertyMappings := GetPropertyMappings(key, context)
	if propertyMappings == nil {
		log.Debugf("No property mapping found for key %v of resource type %v", key, context.ResourceAddress)
		return nil, nil
	}

	var valueFound interface{}
	for _, propertyMapping := range propertyMappings {
		pathProperty := propertyMapping["paths"]
		paths, err := ReadPaths(pathProperty)
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
			valueFounds, err := utils.JsonGet(path, context.Resource)
			if err != nil {
				return nil, err
			}
			if len(valueFounds) == 0 {
				// Try to resolve it against the whole plan
				valueFounds, err = utils.JsonGet(path, *TfPlan)
				if err != nil {
					return nil, err
				}
			}
			if len(valueFounds) > 0 {
				// TODO check if we can safely remove this
				// if len(valueFounds) > 1 {
				// 	return nil, fmt.Errorf("Found more than one value for property %v of resource type %v", key, context.ResourceAddress)
				// }
				valueFound = valueFounds[0]
			}
		}

		if valueFound != nil {
			valueFound, err = ApplyRegex(valueFound, propertyMapping, context.ResourceAddress)
			if err != nil {
				return nil, err
			}
			valueFound, err = ApplyReference(valueFound, propertyMapping, context.ResourceAddress)
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
		defaultValue, err := GetDefaultValue(key, context)
		if err != nil {
			return nil, err
		}

		if defaultValue != nil {
			return defaultValue, nil
		}
	}

	return nil, nil
}

func resolvePlaceholders(input string, context TFContext) (string, error) {
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

func resolvePlaceholder(expression string, context TFContext) (string, error) {
	result := ""
	if strings.HasPrefix(expression, "this.") {
		thisProperty := strings.TrimPrefix(expression, "this.")
		value, err := GetValue(thisProperty, *context.ParentContext)
		if err != nil {
			return "", errors.Wrapf(err, "Cannot get value for variable %s", expression)
		}
		if value == nil {
			return "", errors.Errorf("No value found for variable %s", expression)
		}
		return fmt.Sprintf("%v", value.Value), err
	} else if strings.HasPrefix(expression, "config.") {
		configProperty := strings.TrimPrefix(expression, "config.")
		value := viper.GetFloat64(configProperty)
		return fmt.Sprintf("%v", value), nil
	}
	variable, err := GetVariable(expression, context, context)
	if err != nil {
		return "", err
	}
	if variable != nil {
		result = fmt.Sprintf("%v", variable)
	}
	return result, nil
}

func GetDefaultValue(key string, context TFContext) (*ValueWithUnit, error) {
	propertyMappings := GetPropertyMappings(key, context)
	if propertyMappings == nil {
		log.Debugf("No property mapping found for key %v of resource type %v", key, context.ResourceAddress)
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

		var err error
		valueFound, err = ApplyRegex(valueFound, propertyMapping, context.ResourceAddress)
		if err != nil {
			return nil, err
		}
		valueFound, err = ApplyReference(valueFound, propertyMapping, context.ResourceAddress)
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
	return nil, nil

}

func GetVariable(name string, context TFContext, parentContext TFContext) (interface{}, error) {
	variablesMapping := context.Mapping["variables"]
	if variablesMapping == nil {
		return nil, nil
	}
	if variables, ok := variablesMapping.(map[string]interface{}); ok {
		variableContext := TFContext{
			Resource:        context.Resource,
			Mapping:         variables,
			ResourceAddress: context.ResourceAddress + ".variables",
			ParentContext:   &parentContext,
		}
		value, err := GetValue(name, variableContext)
		if err != nil {
			return nil, err
		}
		if value == nil {
			return nil, fmt.Errorf("Cannot get variable : %v", name)
		}
		return value.Value, nil
	} else {
		return nil, fmt.Errorf("Cannot convert variables to map[string]interface{}: %v", variablesMapping)
	}
}
