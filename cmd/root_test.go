package cmd

import (
	"bytes"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	_ "github.com/carboniferio/carbonifer/internal/testutils"
)

func TestRoot(t *testing.T) {

	actual := new(bytes.Buffer)
	RootCmd.SetOutput(actual)
	if err := RootCmd.Execute(); err != nil {
		log.Debug(err)
	}

	assert.Contains(t, actual.String(), "Usage:", "Default run should return the usage")
}

func TestRootPlan(t *testing.T) {
	wd := "test/terraform/gcp_1"

	b := new(bytes.Buffer)
	RootCmd.SetOutput(b)
	RootCmd.SetArgs([]string{"plan", wd})
	err := RootCmd.Execute()
	if err != nil {
		log.Debug(err)
	}

	assert.True(t, test_planCmdHasRun)

}
