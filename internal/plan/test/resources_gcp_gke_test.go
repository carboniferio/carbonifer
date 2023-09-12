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

func TestGetResource_GCP_GKE(t *testing.T) {
	testutils.SkipWithCreds(t)

	// reset
	terraform.ResetTerraformExec()

	t.Setenv("GOOGLE_OAUTH_ACCESS_TOKEN", "")

	wd := path.Join(testutils.RootDir, "test/terraform/gcp_gke")
	viper.Set("workdir", wd)

	wantResources := map[string]resources.Resource{
		"google_container_cluster.my_cluster": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:         "my_cluster",
				ResourceType: "google_container_cluster",
				Provider:     providers.GCP,
				Region:       "europe-west9",
				Count:        5,
				Address:      "google_container_cluster.my_cluster",
			},
			Specs: &resources.ComputeResourceSpecs{
				VCPUs:             int32(2),
				MemoryMb:          int32(7680),
				ReplicationFactor: 3,
				HddStorage:        decimal.Zero,
				SsdStorage:        decimal.NewFromInt(2725),
			},
		},
		"google_container_cluster.my_cluster_no_pool": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:         "my_cluster_no_pool",
				ResourceType: "google_container_cluster",
				Provider:     providers.GCP,
				Region:       "europe-west9",
				Count:        4,
				Address:      "google_container_cluster.my_cluster_no_pool",
			},
			Specs: &resources.ComputeResourceSpecs{
				VCPUs:             int32(2),
				MemoryMb:          int32(7680),
				ReplicationFactor: 3,
				HddStorage:        decimal.Zero,
				SsdStorage:        decimal.NewFromInt(950),
			},
		},
		"google_container_cluster.auto_provisioned": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:         "auto_provisioned",
				ResourceType: "google_container_cluster",
				Provider:     providers.GCP,
				Region:       "europe-west9",
				Count:        1,
				Address:      "google_container_cluster.auto_provisioned",
			},
			Specs: &resources.ComputeResourceSpecs{
				VCPUs:             int32(5),
				MemoryMb:          int32(10240),
				ReplicationFactor: 3,
				HddStorage:        decimal.Zero,
				SsdStorage:        decimal.NewFromInt(300),
				GpuTypes:          []string{"nvidia-tesla-k80", "nvidia-tesla-k80", "nvidia-tesla-k80"},
			},
		},
		"google_container_cluster.my_cluster_sub_pool": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:         "my_cluster_sub_pool",
				ResourceType: "google_container_cluster",
				Provider:     providers.GCP,
				Region:       "europe-west9",
				Count:        4,
				Address:      "google_container_cluster.my_cluster_sub_pool",
			},
			Specs: &resources.ComputeResourceSpecs{
				VCPUs:             int32(2),
				MemoryMb:          int32(7680),
				ReplicationFactor: 3,
				HddStorage:        decimal.Zero,
				SsdStorage:        decimal.NewFromInt(950),
			},
		},
		"google_container_cluster.my_cluster_autoscaled": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:         "my_cluster_autoscaled",
				ResourceType: "google_container_cluster",
				Provider:     providers.GCP,
				Region:       "europe-west9",
				Count:        12,
				Address:      "google_container_cluster.my_cluster_autoscaled",
			},
			Specs: &resources.ComputeResourceSpecs{
				VCPUs:             int32(2),
				MemoryMb:          int32(7680),
				ReplicationFactor: 3,
				HddStorage:        decimal.Zero,
				SsdStorage:        decimal.NewFromInt(150),
			},
		},
		"google_container_cluster.my_cluster_autoscaled_monozone": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:         "my_cluster_autoscaled_monozone",
				ResourceType: "google_container_cluster",
				Provider:     providers.GCP,
				Region:       "europe-west9",
				Count:        12,
				Address:      "google_container_cluster.my_cluster_autoscaled_monozone",
			},
			Specs: &resources.ComputeResourceSpecs{
				VCPUs:             int32(2),
				MemoryMb:          int32(7680),
				ReplicationFactor: 1,
				HddStorage:        decimal.Zero,
				SsdStorage:        decimal.NewFromInt(150),
			},
		},
		"google_container_cluster.my_cluster_autoscaled_total": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:         "my_cluster_autoscaled_total",
				ResourceType: "google_container_cluster",
				Provider:     providers.GCP,
				Region:       "europe-west9",
				Count:        70,
				Address:      "google_container_cluster.my_cluster_autoscaled_total",
			},
			Specs: &resources.ComputeResourceSpecs{
				VCPUs:             int32(2),
				MemoryMb:          int32(7680),
				ReplicationFactor: 1,
				HddStorage:        decimal.Zero,
				SsdStorage:        decimal.NewFromInt(150),
			},
		},
	}
	tfPlan, err := terraform.TerraformPlan()
	assert.NoError(t, err)
	gotResources, err := plan.GetResources(tfPlan)
	assert.NoError(t, err)
	for _, got := range gotResources {
		if got.GetIdentification().ResourceType == "google_container_node_pool" {
			// This should not exists, it should be ignored
			assert.Fail(t, "google_container_node_pool should be ignored")
		} else if got.GetIdentification().ResourceType == "google_container_cluster" {
			assert.Equal(t, wantResources[got.GetAddress()], got)
		} else {
			// Anything else should be unsupported
			assert.IsType(t, resources.UnsupportedResource{}, got)
		}
	}
}
