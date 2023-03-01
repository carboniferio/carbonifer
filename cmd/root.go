/*
Copyright © 2023 contact@carbonifer.io

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"os"

	"github.com/carboniferio/carbonifer/internal/utils"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "carbonifer",
	Short: "Control carbon emission of your cloud infrastructure",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.carbonifer.yaml)")
	RootCmd.PersistentFlags().StringP("format", "f", "", "format of output ('text' or 'json').\ndefault: 'text'")
	RootCmd.PersistentFlags().StringP("output", "o", "", "output file")
	RootCmd.PersistentFlags().BoolP("debug", "d", false, "print debug logs")
	RootCmd.PersistentFlags().BoolP("info", "i", false, "print info logs")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	utils.InitConfig(cfgFile)

	// Set working directory
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	viper.SetDefault("workdir", currentDir)

	// Set log level from command flags
	info, _ := RootCmd.Flags().GetBool("info")
	debug, _ := RootCmd.Flags().GetBool("debug")
	if info {
		logrus.SetLevel(logrus.InfoLevel)
	} else if debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		viper.SetDefault("log.level", "warning")
	}

	// Bind Viper and Cobra flags
	if err := viper.BindPFlag("out.format", RootCmd.PersistentFlags().Lookup("format")); err != nil {
		log.Panic(err)
	}

	if err := viper.BindPFlag("out.file", RootCmd.PersistentFlags().Lookup("output")); err != nil {
		log.Panic(err)
	}

}
