package terraform_test

import (
	"context"
	"path"
	"testing"

	"github.com/carboniferio/carbonifer/internal/terraform"
	"github.com/carboniferio/carbonifer/internal/testutils"
	_ "github.com/carboniferio/carbonifer/internal/testutils"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetTerraformExec(t *testing.T) {
	// reset
	terraform.ResetTerraformExec()

	viper.Set("workdir", ".")
	tfExec, err := terraform.GetTerraformExec()
	assert.NoError(t, err)
	assert.NotNil(t, tfExec)

}

func TestGetTerraformExec_NotExistingExactVersion(t *testing.T) {
	// reset
	t.Setenv("PATH", "")
	terraform.ResetTerraformExec()

	wantedVersion := "1.2.0"
	viper.Set("workdir", ".")
	viper.Set("terraform.version", wantedVersion)
	terraform.ResetTerraformExec()
	tfExec, err := terraform.GetTerraformExec()
	assert.NoError(t, err)
	assert.NotNil(t, tfExec)
	version, _, err := tfExec.Version(context.Background(), true)
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, version.String(), wantedVersion)

}

func TestGetTerraformExec_NotExistingNoVersion(t *testing.T) {
	// reset
	t.Setenv("PATH", "")
	terraform.ResetTerraformExec()
	viper.Set("terraform.version", "")

	viper.Set("workdir", ".")

	tfExec, err := terraform.GetTerraformExec()
	assert.NoError(t, err)
	assert.NotNil(t, tfExec)
}

func TestTerraformPlan_NoFile(t *testing.T) {
	// reset
	terraform.ResetTerraformExec()

	wd := path.Join(testutils.RootDir, "test/terraform/empty")
	logrus.Infof("workdir: %v", wd)
	viper.Set("workdir", wd)

	_, err := terraform.TerraformPlan()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No configuration files")
}

func TestTerraformPlan_NoTfFile(t *testing.T) {
	// reset
	terraform.ResetTerraformExec()

	wd := path.Join(testutils.RootDir, "test/terraform/notTf")
	logrus.Infof("workdir: %v", wd)
	viper.Set("workdir", wd)

	_, err := terraform.TerraformPlan()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No configuration files")
}

func TestTerraformPlan_BadTfFile(t *testing.T) {
	// reset
	terraform.ResetTerraformExec()

	wd := path.Join(testutils.RootDir, "test/terraform/badTf")
	logrus.Infof("workdir: %v", wd)
	viper.Set("workdir", wd)

	_, err := terraform.TerraformPlan()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "problems")
}

func TestTerraformPlan_MissingCreds(t *testing.T) {
	// reset
	terraform.ResetTerraformExec()

	wd := path.Join(testutils.RootDir, "test/terraform/gcp_images")
	viper.Set("workdir", wd)

	_, err := terraform.TerraformPlan()
	assert.IsType(t, (*terraform.ProviderAuthError)(nil), err)
}

func TestTerraformShow_JSON(t *testing.T) {
	// reset
	terraform.ResetTerraformExec()

	tfPlan, err := terraform.CarboniferPlan("test/terraform/planJson/plan.json")
	assert.NoError(t, err)
	assert.Equal(t, tfPlan.TerraformVersion, "1.3.7")

}

func TestTerraformShow_NotExistJSON(t *testing.T) {
	// reset
	terraform.ResetTerraformExec()

	_, err := terraform.CarboniferPlan("test/terraform/planJson/plan2.json")
	assert.Error(t, err)
}

func TestTerraformShow_RawPlan(t *testing.T) {
	// reset
	terraform.ResetTerraformExec()

	tfPlan, err := terraform.CarboniferPlan("test/terraform/planRaw/plan.tfplan")

	assert.NoError(t, err)
	assert.Equal(t, tfPlan.TerraformVersion, "1.4.6")

}

func TestTerraformShow_WithUnsetVar(t *testing.T) {
	// reset
	terraform.ResetTerraformExec()

	_, err := terraform.CarboniferPlan("test/terraform/planRaw")
	assert.Error(t, err)
	assert.ErrorContains(t, err, "machine_type")

}

func TestTerraformShow_SetVarDifferentFromPlanFile(t *testing.T) {
	// reset
	terraform.ResetTerraformExec()

	t.Setenv("TF_VAR_machine_type", "f1-medium")

	wd := path.Join(testutils.RootDir, "test/terraform/planRaw")
	plan, err := terraform.CarboniferPlan(wd)
	assert.NoError(t, err)
	assert.Equal(t, plan.Variables["machine_type"].Value, "f1-medium")

	wd2 := path.Join(testutils.RootDir, "test/terraform/planRaw/plan.tfplan")
	plan2, err2 := terraform.CarboniferPlan(wd2)
	assert.NoError(t, err2)
	assert.Equal(t, plan2.Variables["machine_type"].Value, "f1-micro")
}
