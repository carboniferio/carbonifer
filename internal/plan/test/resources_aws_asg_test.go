package plan_test

import (
	"path"
	"testing"

	"github.com/carboniferio/carbonifer/internal/plan"
	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/carboniferio/carbonifer/internal/terraform"
	"github.com/carboniferio/carbonifer/internal/testutils"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetResource_AWSASG(t *testing.T) {

	testutils.SkipWithCreds(t)

	// reset
	terraform.ResetTerraformExec()

	wd := path.Join(testutils.RootDir, "test/terraform/aws_asg")
	viper.Set("workdir", wd)

	wantResources := map[string]resources.Resource{
		"aws_autoscaling_group.asg_with_launchconfig": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:              "asg_with_launchconfig",
				Address:           "aws_autoscaling_group.asg_with_launchconfig",
				ResourceType:      "aws_autoscaling_group",
				Provider:          providers.AWS,
				Region:            "eu-west-3",
				Count:             6,
				ReplicationFactor: 1,
			},
			Specs: &resources.ComputeResourceSpecs{
				VCPUs:    int32(4),
				MemoryMb: int32(16384),

				HddStorage: decimal.Zero,
				SsdStorage: decimal.NewFromInt(180),
			},
		},
		"aws_autoscaling_group.asg_launch_template": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:              "asg_launch_template",
				Address:           "aws_autoscaling_group.asg_launch_template",
				ResourceType:      "aws_autoscaling_group",
				Provider:          providers.AWS,
				Region:            "eu-west-3",
				Count:             6,
				ReplicationFactor: 1,
			},
			Specs: &resources.ComputeResourceSpecs{
				VCPUs:    int32(4),
				MemoryMb: int32(16384),

				HddStorage: decimal.NewFromInt(300),
				SsdStorage: decimal.NewFromInt(150),
			},
		},
	}
	tfPlan, err := terraform.TerraformPlan()
	assert.NoError(t, err)
	gotResources, err := plan.GetResources(tfPlan)
	assert.NoError(t, err)
	for _, got := range gotResources {
		if got.GetIdentification().ResourceType == "aws_launch_configuration" {
			// This should not exists, it should be ignored
			assert.Fail(t, "aws_launch_configuration should be ignored")
		} else if got.GetIdentification().ResourceType == "aws_autoscaling_group" {
			assert.Equal(t, wantResources[got.GetAddress()], got)
		} else {
			// Anything else should be unsupported
			assert.IsType(t, resources.UnsupportedResource{}, got)
		}
	}
}
