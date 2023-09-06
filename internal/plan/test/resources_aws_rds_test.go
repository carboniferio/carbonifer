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

func TestGetResource_RDS(t *testing.T) {

	testutils.SkipWithCreds(t)

	// reset
	terraform.ResetTerraformExec()

	wd := path.Join(testutils.RootDir, "test/terraform/aws_rds")
	viper.Set("workdir", wd)

	wantResources := map[string]resources.Resource{
		"aws_db_instance.first": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Address:      "aws_db_instance.first",
				Name:         "first",
				ResourceType: "aws_db_instance",
				Provider:     providers.AWS,
				Region:       "eu-west-3",
				Count:        1,
			},
			Specs: &resources.ComputeResourceSpecs{
				VCPUs:             int32(2),
				MemoryMb:          int32(8192),
				ReplicationFactor: 2,
				HddStorage:        decimal.Zero,
				SsdStorage:        decimal.NewFromInt(300),
			},
		},
		"aws_db_instance.second": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Address:      "aws_db_instance.second",
				Name:         "second",
				ResourceType: "aws_db_instance",
				Provider:     providers.AWS,
				Region:       "eu-west-3",
				Count:        1,
			},
			Specs: &resources.ComputeResourceSpecs{
				VCPUs:             int32(2),
				MemoryMb:          int32(8192),
				ReplicationFactor: 1,
				HddStorage:        decimal.Zero,
				SsdStorage:        decimal.NewFromInt(200),
			},
		},
		"aws_db_instance.third": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Address:      "aws_db_instance.third",
				Name:         "third",
				ResourceType: "aws_db_instance",
				Provider:     providers.AWS,
				Region:       "eu-west-3",
				Count:        1,
			},
			Specs: &resources.ComputeResourceSpecs{
				VCPUs:             int32(2),
				MemoryMb:          int32(8192),
				ReplicationFactor: 1,
				HddStorage:        decimal.Zero,
				SsdStorage:        decimal.NewFromInt(300),
			},
		},
	}
	tfPlan, err := terraform.TerraformPlan()
	assert.NoError(t, err)
	gotResources, err := plan.GetResources(tfPlan)
	assert.NoError(t, err)
	for _, got := range gotResources {
		assert.Equal(t, wantResources[got.GetAddress()], got)
	}
}
