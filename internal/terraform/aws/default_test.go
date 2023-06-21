package aws

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
)

func Test_getDefaultRegion_providerConstant(t *testing.T) {
	awsConfigs := &tfjson.ProviderConfig{
		Name: "aws",
		Expressions: map[string]*tfjson.Expression{
			"region": {
				ExpressionData: &tfjson.ExpressionData{
					ConstantValue: "test1",
				},
			},
		},
	}

	tfPlan := &tfjson.Plan{}

	region := getDefaultRegion(awsConfigs, tfPlan)
	assert.Equal(t, "test1", region)

}

func Test_getDefaultRegion_providerVariable(t *testing.T) {
	awsConfigs := &tfjson.ProviderConfig{
		Name: "aws",
		Expressions: map[string]*tfjson.Expression{
			"region": {
				ExpressionData: &tfjson.ExpressionData{
					References: []string{"var.region"},
				},
			},
		},
	}

	tfPlan := &tfjson.Plan{
		Variables: map[string]*tfjson.PlanVariable{
			"region": {
				Value: "test2",
			},
		},
	}

	region := getDefaultRegion(awsConfigs, tfPlan)
	assert.Equal(t, "test2", region)

}

func Test_getDefaultRegion_EnvVar(t *testing.T) {
	awsConfigs := &tfjson.ProviderConfig{
		Name:        "aws",
		Expressions: map[string]*tfjson.Expression{},
	}

	tfPlan := &tfjson.Plan{}

	t.Setenv("AWS_REGION", "test3")

	region := getDefaultRegion(awsConfigs, tfPlan)
	assert.Equal(t, "test3", region)

}

func Test_getDefaultRegion_EnvDefaultVar(t *testing.T) {
	awsConfigs := &tfjson.ProviderConfig{
		Name:        "aws",
		Expressions: map[string]*tfjson.Expression{},
	}

	tfPlan := &tfjson.Plan{}

	t.Setenv("AWS_DEFAULT_REGION", "test4")

	region := getDefaultRegion(awsConfigs, tfPlan)
	assert.Equal(t, "test4", region)

}

func Test_getDefaultRegion_AWSConfigFile(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create AWS config file
	awsConfigFile := filepath.Join(tmpDir, "config")
	f, err := os.Create(awsConfigFile)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	_, err = io.WriteString(f, "[default]\nregion = region_from_config_file\n")
	if err != nil {
		t.Fatal(err)
	}

	// Set the AWS_SDK_LOAD_CONFIG environment variable
	t.Setenv("AWS_SDK_LOAD_CONFIG", "1")
	t.Setenv("AWS_CONFIG_FILE", awsConfigFile)

	awsConfigs := &tfjson.ProviderConfig{
		Name:        "aws",
		Expressions: map[string]*tfjson.Expression{},
	}

	tfPlan := &tfjson.Plan{}

	region := getDefaultRegion(awsConfigs, tfPlan)
	assert.Equal(t, "region_from_config_file", region)

}
