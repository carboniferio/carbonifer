package testutils

import (
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/carboniferio/carbonifer/internal/utils"
)

var RootDir string

func init() {
	_, filename, _, _ := runtime.Caller(0)
	RootDir = path.Join(path.Dir(filename), "../..")
	err := os.Chdir(RootDir)
	if err != nil {
		panic(err)
	}
	configFile := path.Join(RootDir, "test/config/default_conf.yaml")
	utils.InitWithConfig(configFile)

	// Set fake GCP auth
	os.Setenv("GOOGLE_OAUTH_ACCESS_TOKEN", "foo")

}

func SkipWithCreds(t *testing.T) {
	if os.Getenv("SKIP_WITH_CREDENTIALS") != "" {
		t.Skip("Skipping testing requiring providers credentials")
	}
}
