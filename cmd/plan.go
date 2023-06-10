/*
Copyright © 2023 Carbonifer contact@carbonifer.io
*/
package cmd

import (
	"bufio"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/carboniferio/carbonifer/internal/estimate"
	"github.com/carboniferio/carbonifer/internal/output"
	"github.com/carboniferio/carbonifer/internal/terraform"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var test_planCmdHasRun = false

// planCmd represents the plan command
var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Estimate CO2 from your infrastructure code",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		test_planCmdHasRun = true
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

		fileInfo, err := os.Stat(input)
		if err != nil {
			// Handle error
			panic(err)
		}

		var tfPlan *tfjson.Plan
		// If the path points to a file, run show
		if !fileInfo.IsDir() {
			parentDir := filepath.Dir(input)
			viper.Set("workdir", parentDir)
			tfPlan, err = terraform.TerraformShow(input)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			// If the path points to a directory, run plan
			viper.Set("workdir", input)
			tfPlan, err = terraform.TerraformPlan()
			if err != nil {
				if e, ok := err.(*terraform.ProviderAuthError); ok {
					log.Warnf("Skipping Authentication error: %v", e)
				} else {
					log.Fatal(err)
				}
			}
		}

		// Read resources from terraform plan
		resources, err := terraform.GetResources(tfPlan)
		if err != nil {
			log.Fatal(err)
		}

		// Estimate CO2 emissions
		estimations := estimate.EstimateResources(resources)

		// Generate report
		reportText := ""
		if viper.Get("out.format") == "json" {
			reportText = output.GenerateReportJson(estimations)
		} else {
			reportText = output.GenerateReportText(estimations)
		}

		// Print out report
		outFile := viper.Get("out.file").(string)
		if outFile == "" {
			log.Debug("output : stdout")
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
