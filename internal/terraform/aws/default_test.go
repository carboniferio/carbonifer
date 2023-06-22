package aws

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/carboniferio/carbonifer/internal/testutils"

	"github.com/carboniferio/carbonifer/internal/utils"
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

func Test_getDefaultRegion_ModuleOutput(t *testing.T) {
	awsConfigs := &tfjson.ProviderConfig{
		Name: "aws",
		Expressions: map[string]*tfjson.Expression{
			"region": {
				ExpressionData: &tfjson.ExpressionData{
					References: []string{
						"module.module1.region_output",
						"module.globals"},
				},
			},
		},
	}

	tfPlan := &tfjson.Plan{
		Config: &tfjson.Config{
			RootModule: &tfjson.ConfigModule{
				ModuleCalls: map[string]*tfjson.ModuleCall{
					"module1": {
						Module: &tfjson.ConfigModule{
							Outputs: map[string]*tfjson.ConfigOutput{
								"region_output": {
									Expression: &tfjson.Expression{
										ExpressionData: &tfjson.ExpressionData{
											References: []string{"var.region"},
										},
									},
									Description: "The AWS region to use for resources.",
									Sensitive:   false,
								},
							},
						},
					},
				},
			},
		},
		Variables: map[string]*tfjson.PlanVariable{
			"region": {
				Value: "region_from_module_output",
			},
		},
	}

	region := getDefaultRegion(awsConfigs, tfPlan)
	assert.Equal(t, "region_from_module_output", region)
}

func Test_getDefaultRegion_ModuleVariable(t *testing.T) {
	awsConfigs := &tfjson.ProviderConfig{
		Name: "aws",
		Expressions: map[string]*tfjson.Expression{
			"region": {
				ExpressionData: &tfjson.ExpressionData{
					References: []string{"module.globals.common_region"},
				},
			},
		},
	}

	tfPlan := &tfjson.Plan{
		Config: &tfjson.Config{
			RootModule: &tfjson.ConfigModule{
				ModuleCalls: map[string]*tfjson.ModuleCall{
					"globals": {
						Module: &tfjson.ConfigModule{
							Outputs: map[string]*tfjson.ConfigOutput{
								"common_region": {
									Expression: &tfjson.Expression{
										ExpressionData: &tfjson.ExpressionData{
											References: []string{"var.region"},
										},
									},
									Description: "The AWS region to use for resources.",
								},
							},
							Variables: map[string]*tfjson.ConfigVariable{
								"region": {
									Default: "region_module_variable",
								},
							},
						},
					},
				},
			},
		},
	}

	region := getDefaultRegion(awsConfigs, tfPlan)
	assert.Equal(t, "region_module_variable", region)
}

func TestGetValueOfExpression_ModuleCalls(t *testing.T) {
	plan := utils.LoadPlan("test/terraform/planJson/plan_with_module_calls.json") // Replace with the path to your plan JSON
	expr := &tfjson.Expression{
		ExpressionData: &tfjson.ExpressionData{
			References: []string{"module.module2.module1_region"},
		},
	}

	value, err := utils.GetValueOfExpression(expr, plan)
	assert.NoError(t, err)
	assert.Equal(t, "region_from_module_calls", value)
}
