package output

import (
	"encoding/json"

	"github.com/carboniferio/carbonifer/internal/estimate/estimation"
	log "github.com/sirupsen/logrus"
)

// GenerateReportJSON generates a JSON report from an estimation report
func GenerateReportJSON(estimations estimation.EstimationReport) string {
	log.Debug("Generating JSON report")

	reportTextBytes, err := json.MarshalIndent(estimations, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	return string(reportTextBytes)
}
