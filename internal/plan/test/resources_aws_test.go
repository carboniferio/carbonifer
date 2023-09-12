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

func TestGetResource_DiskFromAMI(t *testing.T) {

	testutils.SkipWithCreds(t)

	// reset
	terraform.ResetTerraformExec()

	wd := path.Join(testutils.RootDir, "test/terraform/aws_ec2")
	viper.Set("workdir", wd)

	wantResources := map[string]resources.Resource{
		"aws_instance.foo": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:              "foo",
				Address:           "aws_instance.foo",
				ResourceType:      "aws_instance",
				Provider:          providers.AWS,
				Region:            "eu-west-3",
				Count:             1,
				ReplicationFactor: 1,
			},
			Specs: &resources.ComputeResourceSpecs{
				VCPUs:    int32(4),
				MemoryMb: int32(16384),

				HddStorage: decimal.NewFromInt(80),
				SsdStorage: decimal.NewFromInt(30),
			},
		},
		"aws_ebs_volume.ebs_volume": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Address:           "aws_ebs_volume.ebs_volume",
				Name:              "ebs_volume",
				ResourceType:      "aws_ebs_volume",
				Provider:          providers.AWS,
				Region:            "eu-west-3",
				Count:             1,
				ReplicationFactor: 1,
			},
			Specs: &resources.ComputeResourceSpecs{
				HddStorage: decimal.Zero,
				SsdStorage: decimal.NewFromInt(100),
			},
		},
		"aws_network_interface.foo": resources.UnsupportedResource{
			Identification: &resources.ResourceIdentification{
				Address:      "aws_network_interface.foo",
				Name:         "foo",
				ResourceType: "aws_network_interface",
				Provider:     providers.AWS,
				Count:        1,
			},
		},
		"aws_subnet.my_subnet": resources.UnsupportedResource{
			Identification: &resources.ResourceIdentification{
				Address:      "aws_subnet.my_subnet",
				Name:         "my_subnet",
				ResourceType: "aws_subnet",
				Provider:     providers.AWS,
				Count:        1,
			},
		},
	}
	tfPlan, err := terraform.TerraformPlan()
	assert.NoError(t, err)
	gotResources, err := plan.GetResources(tfPlan)
	assert.NoError(t, err)
	for _, res := range gotResources {
		assert.Equal(t, wantResources[res.GetAddress()], res)

	}
}
