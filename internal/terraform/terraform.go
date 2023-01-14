package terraform

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var terraformExec *tfexec.Terraform

func GetTerraformExec() (*tfexec.Terraform, error) {
	if terraformExec == nil {
		log.Debugf("Finding or installing terraform exec")
		// Check if terraform is already installed
		execPath, err := exec.LookPath("terraform")
		if err != nil {
			log.Info("Terraform exec not found. Installing...")
			execPath = installTerraform()
		} else {
			log.Info("Using Terraform exec from ", execPath)
		}

		terraformExec, err = tfexec.NewTerraform(viper.GetString("workdir"), execPath)
		if err != nil {
			return nil, err
		}
		version, _, err := terraformExec.Version(context.Background(), true)
		if err != nil {
			log.Fatal(err)
		}

		log.Infof("Using terraform %v", version)
		if err != nil {
			log.Fatalf("error running NewTerraform: %v", err)
		}
	}

	return terraformExec, nil
}

func installTerraform() string {
	var execPath string
	terraformInstallDir := viper.GetString("terraform.path")
	if terraformInstallDir != "" {
		log.Debugf("Terraform install dir configured: %v", terraformInstallDir)
	}
	terraformVersion := viper.GetString("terraform.version")
	ctx := context.Background()
	if terraformVersion != "" {
		log.Debugf("Terraform version configured: %v", terraformVersion)
		installer := &releases.ExactVersion{
			Product:    product.Terraform,
			Version:    version.Must(version.NewVersion(terraformVersion)),
			InstallDir: terraformInstallDir,
		}
		var err error
		execPath, err = installer.Install(ctx)
		if err != nil {
			log.Fatalf("error installing Terraform: %v", err)
		}
	} else {
		log.Debugf("Terraform version not configured, picking latest")
		installer := &releases.LatestVersion{
			Product:    product.Terraform,
			InstallDir: terraformInstallDir,
		}
		var err error
		execPath, err = installer.Install(ctx)
		if err != nil {
			log.Fatalf("error installing Terraform: %v", err)
		}
	}

	log.Infof("Terraform is installed in %v", execPath)
	return execPath
}

func TerraformPlan() (*tfjson.Plan, error) {
	tf, err := GetTerraformExec()
	if err != nil {
		return nil, err
	}

	_, err = tf.Validate(context.Background())
	if err != nil {
		return nil, err
	}

	// Create Temp out plan file
	cfDir, err := os.MkdirTemp(viper.GetString("workdir"), ".carbonifer")
	if err != nil {
		log.Panic(err)
	}
	log.Debugf("Created temporary terraform plan directory %v", cfDir)

	defer func() {
		if err := os.RemoveAll(cfDir); err != nil {
			log.Fatal(err)
		}
	}()

	tfPlanFile, err := ioutil.TempFile(cfDir, "plan-*.tfplan")
	if err != nil {
		log.Fatal(err)
	}

	log.Debugf("Running terraform plan in %v", viper.GetString("workdir"))
	out := tfexec.Out(tfPlanFile.Name())
	_, err = tf.Plan(context.Background(), out)
	if err != nil {
		return nil, err
	}

	tfplan, err := tf.ShowPlanFile(context.Background(), tfPlanFile.Name())
	if err != nil {
		log.Panicf("error running  Terraform Show: %s", err)
	}
	return tfplan, nil
}

func GetResources() []resources.ComputeResource {
	log.Debug("Reading planned resources from Terraform plan")
	tfPlan, err := TerraformPlan()
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("Reading resources from Terraform plan: %d resources", len(tfPlan.PlannedValues.RootModule.Resources))
	var computeResources []resources.ComputeResource
	for _, res := range tfPlan.PlannedValues.RootModule.Resources {
		log.Debugf("Reading resource %v", res.Address)
		if strings.HasPrefix(res.Type, "google") {
			computeResource := GetResource(*res)
			if log.IsLevelEnabled(log.DebugLevel) {
				computeJsonStr := "<RESOURCE TYPE CURRENTLY NOT SUPPORTED>"
				if computeResource != nil {
					computeJson, _ := json.Marshal(computeResource)
					computeJsonStr = string(computeJson)
				}
				log.Debugf("  Compute resource : %v", string(computeJsonStr))
			}
			if computeResource != nil {
				computeResources = append(computeResources, *computeResource)
			}
		}
	}
	return computeResources
}
