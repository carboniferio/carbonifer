package utils

import (
	"encoding/json"
	"os"

	log "github.com/sirupsen/logrus"

	tfjson "github.com/hashicorp/terraform-json"
)

func LoadPlan(planFilePath string) *tfjson.Plan {
	planBytes, err := os.ReadFile(planFilePath)
	if err != nil {
		log.Fatalf("Unable to read JSON plan file: %s", err)
	}

	var plan tfjson.Plan
	err = json.Unmarshal(planBytes, &plan)
	if err != nil {
		log.Fatalf("Unable to unmarshal JSON plan file: %s", err)
	}

	return &plan
}
