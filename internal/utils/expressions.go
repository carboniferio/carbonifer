package utils

import (
	"errors"
	"fmt"
	"strings"

	tfjson "github.com/hashicorp/terraform-json"
)

func GetValueOfExpression(expression *tfjson.Expression, tfPlan *tfjson.Plan, configModuleOptional ...*tfjson.ConfigModule) (interface{}, error) {
	var rootModule *tfjson.ConfigModule
	if len(configModuleOptional) > 0 {
		rootModule = configModuleOptional[0]
	} else {
		if tfPlan.Config != nil && tfPlan.Config.RootModule != nil {
			rootModule = tfPlan.Config.RootModule
		}
	}

	if expression.ConstantValue != nil && fmt.Sprintf("%T", expression.ConstantValue) != "*tfjson.unknownConstantValue" {
		// It's a known value, return it as is
		return expression.ConstantValue, nil
	}

	for _, reference := range expression.References {
		refType, ref := splitModuleReference(reference)
		switch refType {
		case "var":
			// First, check in the plan variables
			if val, ok := tfPlan.Variables[ref]; ok {
				return val.Value, nil
			}

			// If rootModule is not nil, check in the root module and the called module variables
			if rootModule != nil {
				// If not found in plan variables, check in the root module variables
				if moduleVariable, ok := rootModule.Variables[ref]; ok {
					return moduleVariable.Default, nil
				}

				// If not found in root module variables, check in the called module variables
				for _, moduleCall := range rootModule.ModuleCalls {
					if moduleVariable, ok := moduleCall.Module.Variables[ref]; ok {
						return moduleVariable.Default, nil
					}
				}
			}
		case "module":
			if rootModule != nil {
				moduleCall, ok := rootModule.ModuleCalls[ref]
				if ok {
					moduleOutput := strings.Split(reference, ".")
					if len(moduleOutput) >= 3 {
						outputKey := moduleOutput[2]
						output, ok := moduleCall.Module.Outputs[outputKey]
						if ok {
							// Recursive call with the new module config
							return GetValueOfExpression(output.Expression, tfPlan, moduleCall.Module)
						}
					}
				}
			}
		}
	}
	return nil, errors.New("no value found for expression")
}

func splitModuleReference(reference string) (string, string) {
	parts := strings.Split(reference, ".")
	if len(parts) > 1 {
		return parts[0], parts[1]
	}
	return parts[0], ""
}
