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

func TestGetResource_ChildModules(t *testing.T) {

	t.Setenv("GOOGLE_OAUTH_ACCESS_TOKEN", "")

	// reset
	terraform.ResetTerraformExec()

	wd := path.Join(testutils.RootDir, "test/terraform/gcp_large")
	viper.Set("workdir", wd)

	wantResources := map[string]resources.Resource{
		"module.backend.module.db.google_sql_database_instance.instance": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:         "instance",
				ResourceType: "google_sql_database_instance",
				Provider:     providers.GCP,
				Region:       "europe-west9",
				Count:        1,
				Address:      "module.backend.module.db.google_sql_database_instance.instance",
			},
			Specs: &resources.ComputeResourceSpecs{
				VCPUs:             int32(1),
				MemoryMb:          int32(1740),
				ReplicationFactor: 2,
				HddStorage:        decimal.Zero,
				SsdStorage:        decimal.NewFromInt(10),
			},
		},
		"module.backend.module.middleware.module.api_ms.google_compute_instance.cbf-test-vm": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:         "cbf-test-vm",
				ResourceType: "google_compute_instance",
				Provider:     providers.GCP,
				Region:       "europe-west9",
				Count:        1,
				Address:      "module.backend.module.middleware.module.api_ms.google_compute_instance.cbf-test-vm",
			},
			Specs: &resources.ComputeResourceSpecs{
				VCPUs:             int32(12),
				MemoryMb:          int32(87040),
				ReplicationFactor: 1,
				HddStorage:        decimal.NewFromInt(10),
				SsdStorage:        decimal.Zero,
			},
		},
		"module.backend.module.middleware.module.users_ms.google_compute_instance.cbf-test-vm": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:         "cbf-test-vm",
				ResourceType: "google_compute_instance",
				Provider:     providers.GCP,
				Region:       "europe-west9",
				Count:        1,
				Address:      "module.backend.module.middleware.module.users_ms.google_compute_instance.cbf-test-vm",
			},
			Specs: &resources.ComputeResourceSpecs{
				VCPUs:             int32(2),
				MemoryMb:          int32(7680),
				ReplicationFactor: 1,
				HddStorage:        decimal.NewFromInt(10),
				SsdStorage:        decimal.Zero,
			},
		},
		"module.network.google_compute_network.vpc_network": resources.UnsupportedResource{
			Identification: &resources.ResourceIdentification{
				Name:         "vpc_network",
				ResourceType: "google_compute_network",
				Provider:     providers.GCP,
				Count:        1,
				Address:      "module.network.google_compute_network.vpc_network",
			},
		},
		"module.network.google_compute_subnetwork.default": resources.UnsupportedResource{
			Identification: &resources.ResourceIdentification{
				Name:         "default",
				ResourceType: "google_compute_subnetwork",
				Provider:     providers.GCP,
				Count:        1,
				Address:      "module.network.google_compute_subnetwork.default",
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
