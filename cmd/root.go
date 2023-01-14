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

	"github.com/heirko/go-contrib/logrusHelper"
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
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".carbonifer" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".carbonifer")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	viper.ReadInConfig()

	// Setup Logrus
	var logrusConfig = logrusHelper.UnmarshalConfiguration(viper.GetViper().Sub("log")) // Unmarshal configuration from Viper
	logrusHelper.SetConfig(log.StandardLogger(), logrusConfig)                          // for e.g. apply it to logrus default instance

	if viper.ConfigFileUsed() != "" {
		log.Infof("Using config file: %v", viper.ConfigFileUsed())
	}

	// Set log level
	info, _ := RootCmd.Flags().GetBool("info")
	debug, _ := RootCmd.Flags().GetBool("debug")
	if info {
		logrus.SetLevel(logrus.InfoLevel)
	} else if debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		viper.SetDefault("log.level", "warning")
	}

	// Viper default values
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	viper.SetDefault("workdir", currentDir)
	viper.SetDefault("out.format", "text")
	viper.SetDefault("unit.time", "h")      // h or m
	viper.SetDefault("unit.power", "W")     // W or kW
	viper.SetDefault("unit.carbon", "g")    // g or kg
	viper.SetDefault("avg_cpu_use", 0.5)    // g or kg
	viper.SetDefault("out.file", "")        // Path of report file. Default stdout
	viper.SetDefault("data.path", "./data") // Path to data files (provider coefficients...)

	// Bind Viper and Cobra flags
	viper.BindPFlag("out.format", RootCmd.PersistentFlags().Lookup("format"))
	viper.BindPFlag("out.file", RootCmd.PersistentFlags().Lookup("output"))

}
