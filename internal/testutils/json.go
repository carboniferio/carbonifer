package testutils

import (
	"encoding/json"
	"log"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/tidwall/gjson"
)

func TfResourceToJson(stateResource tfjson.StateResource) *gjson.Result {
	bytes, err := json.MarshalIndent(stateResource, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	res := gjson.ParseBytes(bytes)

	return &res
}
