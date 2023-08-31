package terraform

import (
	"github.com/pkg/errors"
)

func getValueOfExpression(expression map[string]interface{}, context *tfContext) (interface{}, error) {

	if expression["constant_value"] != nil {
		// It's a known value, return it as is
		return expression["constant_value"], nil
	}
	if expression["references"] == nil {
		return nil, errors.Errorf("No references found in expression: %v", expression)
	}

	references, ok := expression["references"].([]interface{})
	if !ok {
		return nil, errors.Errorf("References is not an array: %v : %T", expression["references"], expression["references"])
	}

	for _, reference := range references {
		reference, ok := reference.(string)
		if !ok {
			return nil, errors.Errorf("Reference is not a string: %v : %T", reference, reference)
		}

		valueFromConsole, err := runTerraformConsole(reference)
		if err != nil {
			continue
		}
		if valueFromConsole != nil && *valueFromConsole != "" {
			return *valueFromConsole, nil
		}

	}
	return nil, errors.New("no value found for expression")
}
