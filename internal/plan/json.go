package plan

import (
	"strings"

	"github.com/carboniferio/carbonifer/internal/utils"
)

const allResourcesQuery = ".planned_values | .. | objects | select(has(\"resources\")) | .resources[]"

func getJSON(query string, json interface{}) ([]interface{}, error) {

	if strings.HasPrefix(query, "select(") {
		results, err := utils.GetJSON(allResourcesQuery+" | "+query, *TfPlan)
		if len(results) > 0 && err == nil {
			return results, nil
		}
		return nil, err
	}

	if strings.HasPrefix(query, ".configuration") || strings.HasPrefix(query, ".prior_state") || strings.HasPrefix(query, ".planned_values") {
		results, err := utils.GetJSON(query, *TfPlan)
		if len(results) > 0 && err == nil {
			return results, nil
		}
		return nil, err
	}

	results, err := utils.GetJSON(query, json)
	if len(results) > 0 && err == nil {
		return results, nil
	}
	return nil, err
}
