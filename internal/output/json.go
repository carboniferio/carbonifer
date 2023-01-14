package output

import (
	"encoding/json"

	"github.com/carboniferio/carbonifer/internal/estimate"
	log "github.com/sirupsen/logrus"
)

func GenerateReportJson(estimations estimate.EstimationReport) string {
	log.Debug("Generating JSON report")

	reportTextBytes, err := json.MarshalIndent(estimations, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	return string(reportTextBytes)
}
