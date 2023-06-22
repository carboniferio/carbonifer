package testutils

import (
	"encoding/json"

	tfjson "github.com/hashicorp/terraform-json"
)

func TfResourceToJson(resource *tfjson.StateResource) (*map[string]interface{}, error) {
	var result map[string]interface{}
	bytes, err := json.Marshal(resource)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
