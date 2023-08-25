package utils

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJsonGet(t *testing.T) {
	// Sample JSON input for testing
	json := map[string]interface{}{
		"name": "John",
		"cars": []string{"Ford", "BMW", "Fiat"},
		"foo":  []map[string]interface{}{{"bar": "baz"}},
		"values": jsonParse(`
		{
			"advanced_machine_features": [],
			"allow_stopping_for_update": null,
			"attached_disk": [],
			"boot_disk": [
				{
				"auto_delete": true,
				"disk_encryption_key_raw": null,
				"initialize_params": [
					{
					"image": "debian-cloud/debian-11",
					"type": "pd-balanced"
					}
				],
				"mode": "READ_WRITE"
				}
			]
		}`),
	}

	// Test cases with different queries
	testCases := []struct {
		query          string
		expectedResult interface{}
		expectError    bool
	}{
		{
			query:          ".name",
			expectedResult: []interface{}{"John"},
			expectError:    false,
		},
		{
			query:          ".cars",
			expectedResult: []interface{}{[]string{"Ford", "BMW", "Fiat"}},
			expectError:    false,
		},
		{
			query:          ".nonexistent",
			expectedResult: []interface{}{},
			expectError:    false,
		},
		{
			query:          ".foo",
			expectedResult: []interface{}{[]map[string]interface{}{{"bar": "baz"}}},
			expectError:    false,
		},
		{
			query: ".values.boot_disk[].initialize_params",
			expectedResult: []interface{}{
				[]interface{}{
					map[string]interface{}{
						"image": "debian-cloud/debian-11",
						"type":  "pd-balanced",
					},
				},
			},
			expectError: false,
		},
		{
			query:          ".prior_state.values.root_module.resources[] | select(.values.self_link == \"boo\") | .values",
			expectedResult: []interface{}{},
			expectError:    false,
		},
	}

	for _, tc := range testCases {
		result, err := JsonGet(tc.query, json)
		if tc.expectError {
			assert.Error(t, err, "Expected an error for query: %s", tc.query)
		} else {
			assert.NoError(t, err, "Unexpected error for query: %s", tc.query)
			assert.Equal(t, tc.expectedResult, result, "Unexpected result for query: %s", tc.query)
		}
	}
}

func jsonParse(jsonString string) map[string]interface{} {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		panic(err)
	}
	return result
}
