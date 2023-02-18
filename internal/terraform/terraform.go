package terraform

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/carboniferio/carbonifer/internal/terraform/gcp"
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

	log.Debug("Running terraform plan in ", tf.WorkingDir())

	ctx := context.Background()

	// Terraform init
	err = tf.Init(ctx)
	if err != nil {
		return nil, err
	}

	// Terraform Validate
	_, err = tf.Validate(ctx)
	if err != nil {
		return nil, err
	}

	// Create Temp out plan file
	cfDir, err := os.MkdirTemp(tf.WorkingDir(), ".carbonifer")
	if err != nil {
		log.Panic(err)
	}
	log.Debugf("Created temporary terraform plan directory %v", cfDir)

	defer func() {
		if err := os.RemoveAll(cfDir); err != nil {
			log.Fatal(err)
		}
	}()

	tfPlanFile, err := os.CreateTemp(cfDir, "plan-*.tfplan")
	if err != nil {
		log.Fatal(err)
	}

	// Log useful info
	log.Debugf("Using temp terraform plan file %v", tfPlanFile.Name())
	log.Debugf("Running terraform plan in %v", tf.WorkingDir())
	log.Debugf("Running terraform exec %v", tf.ExecPath())

	// Run Terraform Plan with an output file
	out := tfexec.Out(tfPlanFile.Name())
	_, err = tf.Plan(ctx, out)
	if err != nil {
		if strings.Contains(err.Error(), "invalid authentication credentials") ||
			strings.Contains(err.Error(), "No credentials loaded") {
			return nil, &ProviderAuthError{ParentError: err}
		}
		return nil, err
	}

	// Run Terraform Show reading file outputed in step above
	tfplan, err := tf.ShowPlanFile(ctx, tfPlanFile.Name())
	if err != nil {
		log.Infof("error running  Terraform Show: %s", err)
		return nil, err
	}
	return tfplan, nil
}

func GetResources() (map[string]resources.Resource, error) {
	log.Debug("Reading planned resources from Terraform plan")
	tfPlan, err := TerraformPlan()
	if err != nil {
		if e, ok := err.(*ProviderAuthError); ok {
			return nil, e
		} else {
			log.Fatal(err)
		}
	}
	log.Debugf("Reading resources from Terraform plan: %d resources", len(tfPlan.PlannedValues.RootModule.Resources))
	resourcesMap := make(map[string]resources.Resource)
	resourceTemplates := make(map[string]*tfjson.ConfigResource)
	dataResources := make(map[string]resources.DataResource)
	if tfPlan.PriorState != nil {
		for _, priorRes := range tfPlan.PriorState.Values.RootModule.Resources {
			log.Debugf("Reading prior state resources %v", priorRes.Address)
			if priorRes.Mode == "data" {
				if strings.HasPrefix(priorRes.Type, "google") {
					dataResource := gcp.GetDataResource(*priorRes)
					dataResources[dataResource.GetKey()] = dataResource
				}
			}
		}
	}

	// Find template first
	for _, res := range tfPlan.Config.RootModule.Resources {
		log.Debugf("Reading resource %v", res.Address)
		if strings.HasPrefix(res.Type, "google") && strings.HasSuffix(res.Type, "_template") {
			if res.Mode == "managed" {
				resourceTemplates[res.Address] = res
			}
		}
	}

	for _, res := range tfPlan.Config.RootModule.Resources {
		log.Debugf("Reading resource %v", res.Address)
		if strings.HasPrefix(res.Type, "google") && !strings.HasSuffix(res.Type, "_template") {
			if res.Mode == "managed" {
				resource := gcp.GetResource(*res, &dataResources, &resourceTemplates)
				resourcesMap[resource.GetAddress()] = resource
				if log.IsLevelEnabled(log.DebugLevel) {
					computeJsonStr := "<RESOURCE TYPE CURRENTLY NOT SUPPORTED>"
					if resource.IsSupported() {
						computeJson, _ := json.Marshal(resource)
						computeJsonStr = string(computeJson)
					}
					log.Debugf("  Compute resource : %v", string(computeJsonStr))
				}
			}
		}
	}
	return resourcesMap, nil
}
