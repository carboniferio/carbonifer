package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/carboniferio/carbonifer/internal/terraform"
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
		var valueInterpolated interface{}
		switch refType {
		case "local":
			return nil, nil
		case "var":
			// First, check in the plan variables
			if val, ok := tfPlan.Variables[ref]; ok {
				valueInterpolated = val.Value
			}

			// If rootModule is not nil, check in the root module and the called module variables
			if rootModule != nil {
				// If not found in plan variables, check in the root module variables
				if moduleVariable, ok := rootModule.Variables[ref]; ok {
					valueInterpolated = moduleVariable.Default
				}

				// If not found in root module variables, check in the called module variables
				for _, moduleCall := range rootModule.ModuleCalls {
					if moduleVariable, ok := moduleCall.Module.Variables[ref]; ok && valueInterpolated == nil {
						valueInterpolated = moduleVariable.Default
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
							value, err := GetValueOfExpression(output.Expression, tfPlan, moduleCall.Module)
							if err != nil {
								continue
							}
							if value != nil {
								valueInterpolated = value
							}
						}
					}
				}
			}
		}

		// Try to get it from terraform console
		if valueInterpolated == nil {

			valueFromConsole, err := terraform.RunTerraformConsole(reference)
			if err != nil {
				continue
			}
			valueInterpolated = valueFromConsole
		}
		return valueInterpolated, nil
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
