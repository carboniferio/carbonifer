package gcp

import tfjson "github.com/hashicorp/terraform-json"

func GetConstFromConfig(resource *tfjson.ConfigResource, key string) interface{} {
	expr := resource.Expressions[key]
	return GetConstFromExpression(expr)
}

func GetConstFromExpression(expr *tfjson.Expression) interface{} {
	if expr != nil {
		if expr.ConstantValue != nil {
			return expr.ConstantValue
		}
	}
	return nil
}
