package cmd

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/carboniferio/carbonifer/internal/estimate"
	"github.com/carboniferio/carbonifer/internal/output"
	"github.com/carboniferio/carbonifer/internal/plan"
	"github.com/carboniferio/carbonifer/internal/terraform"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var testPlanCmdHasRun = false

// planCmd represents the plan command
var planCmd = &cobra.Command{
	Use: "plan",
	Long: `Estimate CO2 from your infrastructure code.

The 'plan' command optionally takes a single argument:

    directory : 
		- default: current directory
		- directory: a terraform project directory
		- file: a terraform plan file (raw or json)
Example usages:
	carbonifer plan
	carbonifer plan /path/to/terraform/project
	carbonifer plan /path/to/terraform/plan.json
	carbonifer plan /path/to/terraform/plan.tfplan`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		testPlanCmdHasRun = true
		log.Debug("Running command 'plan'")

		workdir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		input := workdir
		if len(args) != 0 {
			input = args[0]
			if !filepath.IsAbs(input) {
				input = filepath.Join(workdir, input)
			}
		}

		// Generate or Read Terraform plan
		tfPlan, err := terraform.CarboniferPlan(input)
		if err != nil {
			log.Fatal(err)
		}

		// Read resources from terraform plan
		resources, err := plan.GetResources(tfPlan)
		if err != nil {
			errW := errors.Wrap(err, "Failed to get resources from terraform plan")
			log.Panic(errW)
		}

		// Estimate CO2 emissions
		estimations := estimate.EstimateResources(resources)

		// Generate report
		reportText := ""
		if viper.Get("out.format") == "json" {
			reportText = output.GenerateReportJSON(estimations)
		} else {
			reportText = output.GenerateReportText(estimations)
		}

		// Print out report
		outFile := viper.Get("out.file").(string)
		if outFile == "" {
			log.Debug("output : stdout")
			cmd.SetOut(os.Stdout)
			cmd.Println(reportText)
		} else {
			log.Debug("output :", outFile)
			f, err := os.Create(outFile)
			if err != nil {
				log.Fatal(err)
			}
			outWriter := bufio.NewWriter(f)
			_, err = outWriter.WriteString(reportText)
			if err != nil {
				log.Fatal(err)
			}
			outWriter.Flush()
		}
	},
}

func init() {
	RootCmd.AddCommand(planCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// planCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// planCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
