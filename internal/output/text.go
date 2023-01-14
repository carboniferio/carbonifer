package output

import (
	"fmt"
	"strings"

	"github.com/carboniferio/carbonifer/internal/estimate"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
)

func GenerateReportText(report estimate.EstimationReport) string {
	log.Debug("Generating text report")
	tableString := &strings.Builder{}
	tableString.WriteString("\n  Average estimation of CO2 emissions per instance: \n\n")

	table := tablewriter.NewWriter(tableString)
	table.SetHeader([]string{"resource type", "name", "emissions"})

	for _, resource := range report.Resources {
		table.Append([]string{
			resource.Resource.GetIndentification().ResourceType,
			resource.Resource.GetIndentification().Name,
			fmt.Sprintf(" %v %v", resource.CarbonEmissions.StringFixed(4), report.Info.UnitCarbonEmissionsTime),
		})
	}

	for _, resource := range report.UnsupportedResources {
		table.Append([]string{
			resource.GetIndentification().ResourceType,
			resource.GetIndentification().Name,
			"unsupported",
		})
	}

	table.SetFooter([]string{"", "Total", fmt.Sprintf(" %v %v", report.Total.CarbonEmissions.StringFixed(4), report.Info.UnitCarbonEmissionsTime)})

	// Format
	table.SetAutoFormatHeaders(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetFooterAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(true)
	table.SetColumnSeparator(" ")
	table.SetCenterSeparator(" ")

	table.Render()
	return tableString.String()
}
