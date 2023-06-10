package terraform_test

import (
	"log"
	"path"
	"testing"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/carboniferio/carbonifer/internal/terraform"
	"github.com/carboniferio/carbonifer/internal/testutils"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetResources(t *testing.T) {

	// reset
	terraform.ResetTerraformExec()

	wd := path.Join(testutils.RootDir, "test/terraform/gcp_1")
	viper.Set("workdir", wd)

	wantResources := map[string]resources.Resource{
		"google_compute_disk.first": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:         "first",
				ResourceType: "google_compute_disk",
				Provider:     providers.GCP,
				Region:       "europe-west9",
				Count:        1,
			},
			Specs: &resources.ComputeResourceSpecs{
				GpuTypes:          nil,
				HddStorage:        decimal.NewFromInt(1024),
				SsdStorage:        decimal.Zero,
				MemoryMb:          0,
				VCPUs:             0,
				CPUType:           "",
				ReplicationFactor: 1,
			},
		},
		"google_compute_instance.first": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:         "first",
				ResourceType: "google_compute_instance",
				Provider:     providers.GCP,
				Region:       "europe-west9",
				Count:        1,
			},
			Specs: &resources.ComputeResourceSpecs{
				HddStorage: decimal.Zero,
				SsdStorage: decimal.NewFromFloat(567).Add(decimal.NewFromFloat(375).Add(decimal.NewFromFloat(375))),
				MemoryMb:   87040,
				VCPUs:      12,
				GpuTypes: []string{
					"testing-custom-data-file", // Default of a2-highgpu-1g"
					"nvidia-tesla-k80",         // Added by user in main.tf
					"nvidia-tesla-k80",         // Added by user in main.tf
				},
				ReplicationFactor: 1,
			},
		},
		"google_compute_instance.second": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:         "second",
				ResourceType: "google_compute_instance",
				Provider:     providers.GCP,
				Region:       "europe-west9",
				Count:        1,
			},
			Specs: &resources.ComputeResourceSpecs{
				GpuTypes:          nil,
				HddStorage:        decimal.NewFromFloat(10),
				SsdStorage:        decimal.Zero,
				MemoryMb:          4098,
				VCPUs:             2,
				CPUType:           "",
				ReplicationFactor: 1,
			},
		},
		"google_compute_network.vpc_network": resources.UnsupportedResource{
			Identification: &resources.ResourceIdentification{
				Name:         "vpc_network",
				ResourceType: "google_compute_network",
				Provider:     providers.GCP,
				Region:       "",
				Count:        1,
			},
		},
		"google_compute_region_disk.regional-first": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:         "regional-first",
				ResourceType: "google_compute_region_disk",
				Provider:     providers.GCP,
				Region:       "europe-west9",
				Count:        1,
			},
			Specs: &resources.ComputeResourceSpecs{
				GpuTypes:          nil,
				HddStorage:        decimal.NewFromInt(1024),
				SsdStorage:        decimal.Zero,
				MemoryMb:          0,
				VCPUs:             0,
				CPUType:           "",
				ReplicationFactor: 2,
			},
		},
		"google_compute_subnetwork.first": resources.UnsupportedResource{
			Identification: &resources.ResourceIdentification{
				Name:         "first",
				ResourceType: "google_compute_subnetwork",
				Provider:     providers.GCP,
				Region:       "europe-west9",
				Count:        1,
			},
		},
		"google_sql_database_instance.instance": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:         "instance",
				ResourceType: "google_sql_database_instance",
				Provider:     providers.GCP,
				Region:       "europe-west9",
				Count:        1,
			},
			Specs: &resources.ComputeResourceSpecs{
				GpuTypes:          nil,
				HddStorage:        decimal.Zero,
				SsdStorage:        decimal.NewFromFloat(10),
				MemoryMb:          15360,
				VCPUs:             4,
				CPUType:           "",
				ReplicationFactor: 2,
			},
		},
	}
	tfPlan, _ := terraform.TerraformPlan()
	resources, _ := terraform.GetResources(tfPlan)
	assert.Equal(t, len(resources), len(wantResources))
	for i, resource := range resources {
		wantResource := wantResources[i]
		assert.Equal(t, wantResource, resource)
	}
}

func TestGetResources_MissingCreds(t *testing.T) {
	// reset
	terraform.ResetTerraformExec()

	wd := path.Join(testutils.RootDir, "test/terraform/gcp_images")
	viper.Set("workdir", wd)

	_, err := terraform.TerraformPlan()
	assert.IsType(t, (*terraform.ProviderAuthError)(nil), err)
}

func TestGetResources_DiskImage(t *testing.T) {
	testutils.SkipWithCreds(t)
	// reset
	terraform.ResetTerraformExec()

	t.Setenv("GOOGLE_OAUTH_ACCESS_TOKEN", "")

	wd := path.Join(testutils.RootDir, "test/terraform/gcp_images")
	viper.Set("workdir", wd)

	wantResources := map[string]resources.Resource{
		"google_compute_disk.diskImage": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:         "diskImage",
				ResourceType: "google_compute_disk",
				Provider:     providers.GCP,
				Region:       "europe-west9",
				Count:        1,
			},
			Specs: &resources.ComputeResourceSpecs{
				GpuTypes:          nil,
				HddStorage:        decimal.New(int64(50), 1),
				SsdStorage:        decimal.Zero,
				MemoryMb:          0,
				VCPUs:             0,
				CPUType:           "",
				ReplicationFactor: 1,
			},
		},
	}

	tfPlan, _ := terraform.TerraformPlan()
	resourceList, err := terraform.GetResources(tfPlan)
	if assert.NoError(t, err) {
		assert.Equal(t, len(wantResources), len(resourceList))
		for i, resource := range resourceList {
			wantResource := wantResources[i]
			log.Println(resource.(resources.ComputeResource).Specs.HddStorage)
			assert.EqualValues(t, wantResource, resource)
		}
	}

}

func TestGetResources_GroupInstance(t *testing.T) {
	// reset
	terraform.ResetTerraformExec()

	t.Setenv("GOOGLE_OAUTH_ACCESS_TOKEN", "")

	wd := path.Join(testutils.RootDir, "test/terraform/gcp_group")
	viper.Set("workdir", wd)

	wantResources := map[string]resources.Resource{
		"google_compute_network.vpc_network": resources.UnsupportedResource{
			Identification: &resources.ResourceIdentification{
				Name:         "vpc_network",
				ResourceType: "google_compute_network",
				Provider:     providers.GCP,
				Region:       "",
				Count:        1,
			},
		},
		"google_compute_subnetwork.first": resources.UnsupportedResource{
			Identification: &resources.ResourceIdentification{
				Name:         "first",
				ResourceType: "google_compute_subnetwork",
				Provider:     providers.GCP,
				Region:       "europe-west9",
				Count:        1,
			},
		},
		"google_compute_instance_group_manager.my-group-manager": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:         "my-group-manager",
				ResourceType: "google_compute_instance_group_manager",
				Provider:     providers.GCP,
				Region:       "europe-west9",
				Count:        5,
			},
			Specs: &resources.ComputeResourceSpecs{
				GpuTypes:          nil,
				HddStorage:        decimal.NewFromFloat(20),
				SsdStorage:        decimal.Zero,
				MemoryMb:          8192,
				VCPUs:             2,
				CPUType:           "",
				ReplicationFactor: 1,
			},
		},
	}

	tfPlan, _ := terraform.TerraformPlan()
	resources, err := terraform.GetResources(tfPlan)
	if assert.NoError(t, err) {
		for i, resource := range resources {
			wantResource := wantResources[i]
			assert.EqualValues(t, wantResource, resource)
		}
	}

}

func TestGetResources_InstanceFromTemplate(t *testing.T) {
	// reset
	terraform.ResetTerraformExec()

	t.Setenv("GOOGLE_OAUTH_ACCESS_TOKEN", "")

	wd := path.Join(testutils.RootDir, "test/terraform/gcp_cit")
	viper.Set("workdir", wd)

	wantResources := map[string]resources.Resource{
		"google_compute_network.vpc_network": resources.UnsupportedResource{
			Identification: &resources.ResourceIdentification{
				Name:         "vpc_network",
				ResourceType: "google_compute_network",
				Provider:     providers.GCP,
				Region:       "",
				Count:        1,
			},
		},
		"google_compute_subnetwork.first": resources.UnsupportedResource{
			Identification: &resources.ResourceIdentification{
				Name:         "first",
				ResourceType: "google_compute_subnetwork",
				Provider:     providers.GCP,
				Region:       "europe-west9",
				Count:        1,
			},
		},
		"google_compute_instance_from_template.ifromtpl": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:         "ifromtpl",
				ResourceType: "google_compute_instance_from_template",
				Provider:     providers.GCP,
				Region:       "europe-west9",
				Count:        1,
			},
			Specs: &resources.ComputeResourceSpecs{
				GpuTypes:          nil,
				HddStorage:        decimal.NewFromFloat(20),
				SsdStorage:        decimal.Zero,
				MemoryMb:          8192,
				VCPUs:             2,
				CPUType:           "",
				ReplicationFactor: 1,
			},
		},
	}

	tfPlan, _ := terraform.TerraformPlan()
	resources, err := terraform.GetResources(tfPlan)
	if assert.NoError(t, err) {
		for i, resource := range resources {
			wantResource := wantResources[i]
			assert.EqualValues(t, wantResource, resource)
		}
	}

}
