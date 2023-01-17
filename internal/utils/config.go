package utils

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"gopkg.in/yaml.v3"
)

func LoadViperDefaults() {
	defaultConfigFile, err := os.ReadFile("internal/utils/defaults.yaml")
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Fatal(err)
	}

	defaults := make(map[string]interface{})

	err = yaml.Unmarshal(defaultConfigFile, &defaults)
	if err != nil {
		log.Fatal(err)
	}

	for k, v := range defaults {
		viper.SetDefault(k, v)
	}
	settings := viper.AllSettings()

	log.Debug(settings)
}
