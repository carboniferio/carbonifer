package plan_test

import (
	"log"
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

func TestGetResource_DiskFromAMI(t *testing.T) {

	testutils.SkipWithCreds(t)

	// reset
	terraform.ResetTerraformExec()

	wd := path.Join(testutils.RootDir, "test/terraform/aws_ec2")
	viper.Set("workdir", wd)

	wantResources := map[string]resources.Resource{
		"aws_instance.foo": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:         "foo",
				ResourceType: "aws_instance",
				Provider:     providers.AWS,
				Region:       "eu-west-3",
				Count:        1,
			},
			Specs: &resources.ComputeResourceSpecs{
				VCPUs:             int32(2),
				MemoryMb:          int32(8192),
				ReplicationFactor: 1,
				HddStorage:        decimal.NewFromInt(20),
				SsdStorage:        decimal.NewFromInt(90),
			},
		},
		"aws_ebs_volume.ebs_volume": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:         "ebs_volume",
				ResourceType: "aws_ebs_volume",
				Provider:     providers.AWS,
				Region:       "eu-west-3",
				Count:        1,
			},
			Specs: &resources.ComputeResourceSpecs{
				HddStorage: decimal.Zero,
				SsdStorage: decimal.NewFromInt(100),
			},
		},
	}
	log.Default().Println(wantResources)

	tfPlan, err := terraform.TerraformPlan()
	assert.NoError(t, err)
	gotResources, err := plan.GetResources(tfPlan)
	assert.NoError(t, err)
	for _, res := range gotResources {
		if res.GetIdentification().ResourceType == "aws_instance" {
			assert.Equal(t, wantResources["aws_instance.foo"], res)
		}
		if res.GetIdentification().ResourceType == "aws_ebs_volume" {
			assert.Equal(t, wantResources["aws_ebs_volume.ebs_volume"], res)
		}
	}
}
