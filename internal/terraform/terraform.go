package terraform

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
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

func ResetTerraformExec() {
	terraformExec = nil
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

func terraformInit() (*tfexec.Terraform, *context.Context, error) {
	tf, err := GetTerraformExec()
	if err != nil {
		return nil, nil, err
	}

	log.Debug("Running terraform init in ", tf.WorkingDir())

	ctx := context.Background()

	// Terraform init
	err = tf.Init(ctx)
	if err != nil {
		return nil, &ctx, err
	}

	return tf, &ctx, err
}

// CarboniferPlan generates a Terraform plan from a tfplan file or a Terraform directory
func CarboniferPlan(input string) (*map[string]interface{}, error) {
	fileInfo, err := os.Stat(input)
	if err != nil {
		return nil, err
	}

	// If the path points to a file, run show
	if !fileInfo.IsDir() {
		parentDir := filepath.Dir(input)
		fileName := filepath.Base(input)
		viper.Set("workdir", parentDir)
		tfPlan, err := terraformShow(fileName)
		return tfPlan, err
	}
	// If the path points to a directory, run plan
	viper.Set("workdir", input)
	tfPlan, err := TerraformPlan()
	if err != nil {
		if e, ok := err.(*ProviderAuthError); ok {
			log.Warnf("Skipping Authentication error: %v", e)
		} else {
			return nil, err
		}
	}
	return tfPlan, err

}

func TerraformPlan() (*map[string]interface{}, error) {
	tf, ctx, err := terraformInit()
	if err != nil {
		return nil, err
	}

	// Terraform Validate
	_, err = tf.Validate(*ctx)
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
	err = terraformPlanExec(*ctx, tf, tfPlanFile)
	if err != nil {
		return nil, err
	}

	// Run Terraform Show reading file outputed in step above
	tfplan, err := tf.ShowPlanFile(*ctx, tfPlanFile.Name())
	if err != nil {
		log.Infof("error running  Terraform Show: %s", err)
		return nil, err
	}
	var bytes []byte
	bytes, err = json.MarshalIndent(tfplan, "", "  ")
	if err != nil {
		return nil, err
	}

	var tfplanJSON map[string]interface{}
	err = json.Unmarshal(bytes, &tfplanJSON)
	if err != nil {
		return nil, err
	}
	return &tfplanJSON, nil
}

func terraformPlanExec(ctx context.Context, tf *tfexec.Terraform, tfPlanFile *os.File) error {
	out := tfexec.Out(tfPlanFile.Name())
	_, err := tf.Plan(ctx, out)
	var authError ProviderAuthError
	if err != nil {
		uwErr := err.Error()
		if strings.Contains(uwErr, "invalid authentication credentials") ||
			strings.Contains(uwErr, "No credentials loaded") ||
			strings.Contains(uwErr, "no valid credential") {
			authError = ProviderAuthError{ParentError: err}
			return &authError
		}
		log.Errorf("error running  Terraform Plan: %s", err)
		return err

	}
	return nil
}

func terraformShow(fileName string) (*map[string]interface{}, error) {
	if strings.HasSuffix(fileName, ".json") {
		planFilePath := filepath.Join(viper.GetString("workdir"), fileName)
		log.Debugf("Reading Terraform plan from %v", planFilePath)
		jsonFile, err := os.Open(planFilePath)
		if err != nil {
			return nil, err
		}
		defer jsonFile.Close()
		byteValue, _ := os.ReadFile(planFilePath)

		var tfplan map[string]interface{}
		err = json.Unmarshal(byteValue, &tfplan)
		if err != nil {
			return nil, err
		}
		return &tfplan, nil
	}

	tf, ctx, err := terraformInit()
	if err != nil {
		return nil, err
	}

	// Run Terraform Show
	tfPlan, err := tf.ShowPlanFile(*ctx, fileName)
	if err != nil {
		return nil, err
	}
	tfPlanJSONBytes, err := json.MarshalIndent(tfPlan, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal plan: %v", err)
	}

	var tfPlanJSON map[string]interface{}
	err = json.Unmarshal(tfPlanJSONBytes, &tfPlanJSON)
	if err != nil {
		return nil, err
	}

	return &tfPlanJSON, nil
}

func RunTerraformConsole(command string) (*string, error) {
	tfExec, err := GetTerraformExec()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(tfExec.ExecPath(), "console")

	cmd.Dir = terraformExec.WorkingDir() // set the working directory

	var stdin bytes.Buffer
	stdin.Write([]byte(command + "\n"))
	cmd.Stdin = &stdin

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error running terraform console: %w\nstderr: %s", err, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())

	// Parse the output as JSON
	var result map[string]string
	if err := json.Unmarshal([]byte(output), &result); err == nil {
		// If the output is valid JSON, extract the value of the command key
		var ok bool
		output, ok = result[command]
		if !ok {
			return nil, fmt.Errorf("output does not contain key %q", command)
		}
	}
	// remove quotes surrounding the value
	output = strings.Trim(output, "\"")
	return &output, nil
}
