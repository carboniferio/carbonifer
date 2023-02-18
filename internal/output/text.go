package output

import (
	"fmt"
	"sort"
	"strings"

	"github.com/carboniferio/carbonifer/internal/estimate/estimation"
	"github.com/olekukonko/tablewriter"
	log "github.com/sirupsen/logrus"
)

func GenerateReportText(report estimation.EstimationReport) string {
	log.Debug("Generating text report")
	tableString := &strings.Builder{}
	tableString.WriteString("\n  Average estimation of CO2 emissions per instance: \n\n")

	table := tablewriter.NewWriter(tableString)
	table.SetHeader([]string{"resource type", "name", "count", "emissions per instance"})

	// Default sort
	estimations := report.Resources
	sort.Slice(estimations, func(i, j int) bool {
		return estimations[i].Resource.GetIdentification().Name < estimations[j].Resource.GetIdentification().Name
	})

	for _, resource := range report.Resources {
		table.Append([]string{
			resource.Resource.GetIdentification().ResourceType,
			resource.Resource.GetIdentification().Name,
			fmt.Sprintf("%v", resource.Count),
			fmt.Sprintf(" %v %v", resource.CarbonEmissions.StringFixed(4), report.Info.UnitCarbonEmissionsTime),
		})
	}

	for _, resource := range report.UnsupportedResources {
		table.Append([]string{
			resource.GetIdentification().ResourceType,
			resource.GetIdentification().Name,
			"",
			"unsupported",
		})
	}

	table.SetFooter([]string{"", "Total", report.Total.ResourcesCount.String(), fmt.Sprintf(" %v %v", report.Total.CarbonEmissions.StringFixed(4), report.Info.UnitCarbonEmissionsTime)})

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
