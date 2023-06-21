package utils

import (
	"errors"
	"fmt"
	"strings"

	tfjson "github.com/hashicorp/terraform-json"
)

func GetValueOfExpression(expression *tfjson.Expression, tfPlan *tfjson.Plan) (interface{}, error) {
	if fmt.Sprintf("%T", expression.ConstantValue) != "*tfjson.unknownConstantValue" && expression.ConstantValue != nil {
		// It's a known value, return it as is
		return expression.ConstantValue, nil
	}

	// Constant value is not set or unknown, look up references
	for _, reference := range expression.References {
		ref := strings.TrimPrefix(reference, "var.")
		if val, ok := tfPlan.Variables[ref]; ok {
			return val.Value, nil
		}
	}

	// No variables were found
	return nil, errors.New("no value found for expression")

}
