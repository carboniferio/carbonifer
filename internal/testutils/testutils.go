package testutils

import (
	"log"
	"os"
	"path"
	"runtime"

	"github.com/spf13/viper"
)

var RootDir string

func init() {
	_, filename, _, _ := runtime.Caller(0)
	RootDir = path.Join(path.Dir(filename), "../..")
	err := os.Chdir(RootDir)
	if err != nil {
		panic(err)
	}
	viper.SetConfigFile("test/config/default_conf.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

}
