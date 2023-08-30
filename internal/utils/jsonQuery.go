package utils

import (
	"strings"

	"github.com/itchyny/gojq"
)

// GetJSON returns the result of a jq query on a json object
func GetJSON(query string, json interface{}) ([]interface{}, error) {
	queryParsed, err := gojq.Parse(query)
	if err != nil {
		return nil, err
	}
	iter := queryParsed.Run(json)
	results := []interface{}{}
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			errMsg := err.Error()
			if strings.Contains(errMsg, "annot iterate over: null") {
				continue
			} else {
				return nil, err
			}
		}
		if v != nil {
			results = append(results, v)
		}
	}

	return results, nil
}
