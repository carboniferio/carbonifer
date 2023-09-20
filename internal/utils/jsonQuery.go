package utils

import (
	"fmt"
	"log"
	"strings"

	"github.com/itchyny/gojq"
)

type moduleLoader struct{}

func (*moduleLoader) LoadModule(name string) (*gojq.Query, error) {
	switch name {
	case "carbonifer":
		return gojq.Parse(`
			module { name: "carbonifer" };

			def all_select(a; b):
				.planned_values | .. | objects | select(has("resources")) | .resources[] | select(.[a] == b);

			def extract_disk_key:
				if test("^/dev/sd[a-z]+") or test("^/dev/xvd[a-z]+") then
				  capture("^/dev/(?:sd|xvd)(?<letter>[a-z]+)").letter
				elif test("^/dev/nvme[0-2]?[0-9]n") then
				  capture("^/dev/nvme(?<number>[0-2]?[0-9])n").number | tonumber | (96 + .) | [.] | implode
				else
				  "Unknown format"
				end;
		`)
	}
	return nil, fmt.Errorf("module not found: %q", name)
}

// GetJSON returns the result of a jq query on a json object
func GetJSON(query string, json interface{}) ([]interface{}, error) {
	queryImport := fmt.Sprintf(`import "carbonifer" as cbf; %s`, query)
	queryParsed, err := gojq.Parse(queryImport)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	code, err := gojq.Compile(
		queryParsed, *getGoJQWithModules(),
	)
	if err != nil {
		return nil, err
	}

	iter := code.Run(json)
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

var goJqWithModules *gojq.CompilerOption

func getGoJQWithModules() *gojq.CompilerOption {
	if goJqWithModules == nil {
		goJqWithModulesObj := gojq.WithModuleLoader(&moduleLoader{})
		goJqWithModules = &goJqWithModulesObj
	}
	return goJqWithModules
}
