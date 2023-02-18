package output

import (
	"encoding/json"

	"github.com/carboniferio/carbonifer/internal/estimate/estimation"
	log "github.com/sirupsen/logrus"
)

func GenerateReportJson(estimations estimation.EstimationReport) string {
	log.Debug("Generating JSON report")

	reportTextBytes, err := json.MarshalIndent(estimations, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	return string(reportTextBytes)
}
