package terraform

import (
	"context"
	"path"
	"testing"

	"github.com/carboniferio/carbonifer/internal/testutils"
	_ "github.com/carboniferio/carbonifer/internal/testutils"
	"github.com/carboniferio/carbonifer/internal/utils"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetTerraformExec(t *testing.T) {
	// reset
	ResetTerraformExec()

	viper.Set("workdir", ".")
	tfExec, err := getTerraformExec()
	assert.NoError(t, err)
	assert.NotNil(t, tfExec)

}

func TestGetTerraformExec_NotExistingExactVersion(t *testing.T) {
	// reset
	t.Setenv("PATH", "")
	ResetTerraformExec()

	wantedVersion := "1.2.0"
	viper.Set("workdir", ".")
	viper.Set("terraform.version", wantedVersion)
	ResetTerraformExec()
	tfExec, err := getTerraformExec()
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
	ResetTerraformExec()
	viper.Set("terraform.version", "")

	viper.Set("workdir", ".")

	tfExec, err := getTerraformExec()
	assert.NoError(t, err)
	assert.NotNil(t, tfExec)
}

func TestTerraformPlan_NoFile(t *testing.T) {
	// reset
	ResetTerraformExec()

	wd := path.Join(testutils.RootDir, "test/terraform/empty")
	logrus.Infof("workdir: %v", wd)
	viper.Set("workdir", wd)

	_, err := TerraformPlan()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No configuration files")
}

func TestTerraformPlan_NoTfFile(t *testing.T) {
	// reset
	ResetTerraformExec()

	wd := path.Join(testutils.RootDir, "test/terraform/notTf")
	logrus.Infof("workdir: %v", wd)
	viper.Set("workdir", wd)

	_, err := TerraformPlan()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No configuration files")
}

func TestTerraformPlan_BadTfFile(t *testing.T) {
	// reset
	ResetTerraformExec()

	wd := path.Join(testutils.RootDir, "test/terraform/badTf")
	logrus.Infof("workdir: %v", wd)
	viper.Set("workdir", wd)

	_, err := TerraformPlan()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "problems")
}

func TestTerraformPlan_MissingCreds(t *testing.T) {
	// reset
	ResetTerraformExec()

	wd := path.Join(testutils.RootDir, "test/terraform/gcp_images")
	viper.Set("workdir", wd)

	_, err := TerraformPlan()
	assert.IsType(t, (*ProviderAuthError)(nil), err)
}

func TestTerraformShow_JSON(t *testing.T) {
	// reset
	ResetTerraformExec()

	tfPlan, err := CarboniferPlan("test/terraform/planJson/plan.json")
	assert.NoError(t, err)
	tfVersion, _ := utils.GetJSON(".terraform_version", *tfPlan)
	assert.Equal(t, "1.3.7", tfVersion[0])

}

func TestTerraformShow_NotExistJSON(t *testing.T) {
	// reset
	ResetTerraformExec()

	_, err := CarboniferPlan("test/terraform/planJson/plan2.json")
	assert.Error(t, err)
}

func TestTerraformShow_RawPlan(t *testing.T) {
	// reset
	ResetTerraformExec()

	tfPlan, err := CarboniferPlan("test/terraform/planRaw/plan.tfplan")
	assert.NoError(t, err)
	tfVersion, _ := utils.GetJSON(".terraform_version", *tfPlan)
	assert.Equal(t, tfVersion[0], "1.4.6")

}

func TestTerraformShow_WithUnsetVar(t *testing.T) {
	// reset
	ResetTerraformExec()

	_, err := CarboniferPlan("test/terraform/planRaw")
	assert.Error(t, err)
	assert.ErrorContains(t, err, "machine_type")

}

func TestTerraformShow_SetVarDifferentFromPlanFile(t *testing.T) {
	// reset
	ResetTerraformExec()

	t.Setenv("TF_VAR_machine_type", "f1-medium")

	wd := path.Join(testutils.RootDir, "test/terraform/planRaw")
	plan, err := CarboniferPlan(wd)
	assert.NoError(t, err)
	log.Info(plan)
	machineTypeVar, _ := utils.GetJSON(".variables.machine_type.value", *plan)
	assert.Equal(t, "f1-medium", machineTypeVar[0])

	wd2 := path.Join(testutils.RootDir, "test/terraform/planRaw/plan.tfplan")
	plan2, err2 := CarboniferPlan(wd2)
	assert.NoError(t, err2)
	machineTypeVar2, _ := utils.GetJSON(".variables.machine_type.value", *plan2)
	assert.Equal(t, "f1-micro", machineTypeVar2[0])
}
