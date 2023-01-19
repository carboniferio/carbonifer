package testutils

import (
	"log"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/carboniferio/carbonifer/internal/utils"
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
	utils.LoadViperDefaults()
	viper.AddConfigPath(path.Join(RootDir, "test/config"))
	viper.SetConfigName("default_conf")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	// Set fake GCP auth
	os.Setenv("GOOGLE_OAUTH_ACCESS_TOKEN", "foo")

}

func SkipWithCreds(t *testing.T) {
	if os.Getenv("SKIP_WITH_CREDENTIALS") != "" {
		t.Skip("Skipping testing requiring providers credentials")
	}
}
