package utils

import (
	_ "embed"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"

	"github.com/carboniferio/carbonifer/internal/estimate/estimation"
	"github.com/heirko/go-contrib/logrusHelper"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"gopkg.in/yaml.v3"
)

func InitWithDefaultConfig() {
	initViper("")
	initLogger()
	checkDataConfig()
}

func InitWithConfig(customConfigFilePath string) {
	initViper(customConfigFilePath)
	initLogger()
	checkDataConfig()
}

//go:embed defaults.yaml
var defaultConfigFile []byte

func loadViperDefaults() {
	var defaults map[string]interface{}

	err := yaml.Unmarshal(defaultConfigFile, &defaults)
	if err != nil {
		log.Fatal(err)
	}

	for k, v := range defaults {
		viper.SetDefault(k, v)
	}
	settings := viper.AllSettings()

	log.Debug(settings)
}

func BasePath() string {
	_, b, _, _ := runtime.Caller(0)
	d := filepath.Dir(b)
	return filepath.Join(d, "../..")
}

func initViper(configFilePath string) {
	loadViperDefaults()

	if configFilePath != "" {
		// Use config file from the flag.
		viper.SetConfigFile(configFilePath)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(path.Join(home, ".carbonifer"))
		viper.AddConfigPath("/etc/carbonifer/")
		viper.AddConfigPath("./.carbonifer")
		if viper.ConfigFileUsed() == "" {
			viper.SetConfigType("yaml")
			viper.SetConfigName("config")
		}
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Panic(err)
		}
	}

	// Set absolute data directory
	dataPath := viper.GetString("data.path")
	if dataPath != "" && !filepath.IsAbs(dataPath) {
		basedir := BasePath()
		dataPath = filepath.Join(basedir, dataPath)
	}
	viper.Set("data.path", dataPath)

	if viper.ConfigFileUsed() != "" {
		log.Infof("Using config file: %v", viper.ConfigFileUsed())
	}
}

func initLogger() {
	// Setup Logrus
	logConf := viper.GetViper().Sub("log")
	if logConf != nil {
		var logrusConfig = logrusHelper.UnmarshalConfiguration(logConf) // Unmarshal configuration from Viper
		err := logrusHelper.SetConfig(log.StandardLogger(), logrusConfig)
		if err != nil {
			log.Panic(err)
		}
	}
}

func checkDataConfig() {
	dataPath := viper.GetString("data.path")
	if dataPath != "" {
		path, err := filepath.Abs(dataPath)
		if err != nil {
			log.Fatal(err)
		}
		f, err := os.Open(dataPath)
		if err != nil {
			log.Fatalf("Cannot read data directory \"%v\": %v", path, err)
		}
		defer f.Close()

		_, err = f.Readdirnames(1)
		if err == io.EOF {
			log.Fatalf("Empty data directory \"%v\": %v", path, err)
		}
	}
}

func SortEstimations(resources *[]estimation.EstimationResource) {
	sort.Slice(*resources, func(i, j int) bool {
		return (*resources)[i].Resource.GetAddress() < (*resources)[j].Resource.GetAddress()
	})
}
