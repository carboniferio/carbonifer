package plan

import (
	"strings"

	"github.com/carboniferio/carbonifer/internal/utils"
)

func getJSON(query string, json interface{}) ([]interface{}, error) {

	if strings.Contains(query, "all_select(") {
		results, err := utils.GetJSON(query, *TfPlan)
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
