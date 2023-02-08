package terraform

import (
	"context"
	"path"
	"testing"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/carboniferio/carbonifer/internal/testutils"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetTerraformExec(t *testing.T) {
	// reset
	terraformExec = nil

	viper.Set("workdir", ".")
	tfExec, err := GetTerraformExec()
	assert.NoError(t, err)
	assert.NotNil(t, tfExec)

}

func TestGetTerraformExec_NotExistingExactVersion(t *testing.T) {
	// reset
	t.Setenv("PATH", "")
	terraformExec = nil

	wantedVersion := "1.2.0"
	viper.Set("workdir", ".")
	viper.Set("terraform.version", wantedVersion)
	terraformExec = nil
	tfExec, err := GetTerraformExec()
	assert.NoError(t, err)
	assert.NotNil(t, tfExec)
	version, _, err := tfExec.Version(context.Background(), true)
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, version.String(), wantedVersion)

}

func TestGetTerraformExec_NotExistingNoVersion(t *testing.T) {
	// reset
	t.Setenv("PATH", "")
	terraformExec = nil
	viper.Set("terraform.version", "")

	viper.Set("workdir", ".")

	tfExec, err := GetTerraformExec()
	assert.NoError(t, err)
	assert.NotNil(t, tfExec)
}

func TestTerraformPlan_NoFile(t *testing.T) {
	// reset
	terraformExec = nil

	wd := path.Join(testutils.RootDir, "test/terraform/empty")
	logrus.Infof("workdir: %v", wd)
	viper.Set("workdir", wd)

	_, err := TerraformPlan()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No configuration files")
}

func TestTerraformPlan_NoTfFile(t *testing.T) {
	// reset
	terraformExec = nil

	wd := path.Join(testutils.RootDir, "test/terraform/notTf")
	logrus.Infof("workdir: %v", wd)
	viper.Set("workdir", wd)

	_, err := TerraformPlan()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "No configuration files")
}

func TestTerraformPlan_BadTfFile(t *testing.T) {
	// reset
	terraformExec = nil

	wd := path.Join(testutils.RootDir, "test/terraform/badTf")
	logrus.Infof("workdir: %v", wd)
	viper.Set("workdir", wd)

	_, err := TerraformPlan()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configuration is invalid")
}

func TestGetResources(t *testing.T) {
	// reset
	terraformExec = nil

	wd := path.Join(testutils.RootDir, "test/terraform/gcp_1")
	viper.Set("workdir", wd)

	wantResources := []resources.Resource{
		resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:         "first",
				ResourceType: "google_compute_disk",
				Provider:     providers.GCP,
				Region:       "europe-west9",
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
		resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:         "first",
				ResourceType: "google_compute_instance",
				Provider:     providers.GCP,
				Region:       "europe-west9",
			},
			Specs: &resources.ComputeResourceSpecs{
				HddStorage: decimal.Zero,
				SsdStorage: decimal.NewFromFloat(567).Add(decimal.NewFromFloat(375).Add(decimal.NewFromFloat(375))),
				MemoryMb:   87040,
				VCPUs:      12,
				GpuTypes: []string{
					"nvidia-tesla-a100", // Default of a2-highgpu-1g"
					"nvidia-tesla-k80",  // Added by user in main.tf
					"nvidia-tesla-k80",  // Added by user in main.tf
				},
			},
		},
		resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:         "second",
				ResourceType: "google_compute_instance",
				Provider:     providers.GCP,
				Region:       "europe-west9",
			},
			Specs: &resources.ComputeResourceSpecs{
				GpuTypes:   nil,
				HddStorage: decimal.NewFromFloat(10),
				SsdStorage: decimal.Zero,
				MemoryMb:   4098,
				VCPUs:      2,
				CPUType:    "",
			},
		},
		resources.UnsupportedResource{
			Identification: &resources.ResourceIdentification{
				Name:         "vpc_network",
				ResourceType: "google_compute_network",
				Provider:     providers.GCP,
				Region:       "",
			},
		},
		resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:         "regional-first",
				ResourceType: "google_compute_region_disk",
				Provider:     providers.GCP,
				Region:       "europe-west9",
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
		resources.UnsupportedResource{
			Identification: &resources.ResourceIdentification{
				Name:         "first",
				ResourceType: "google_compute_subnetwork",
				Provider:     providers.GCP,
				Region:       "europe-west9",
			},
		},
	}

	resources, _ := GetResources()
	assert.Equal(t, len(resources), len(wantResources))
	for i, resource := range resources {
		wantResource := wantResources[i]
		assert.Equal(t, wantResource, resource)
	}
}

func TestGetResources_MissingCreds(t *testing.T) {
	// reset
	terraformExec = nil

	wd := path.Join(testutils.RootDir, "test/terraform/gcp_images")
	viper.Set("workdir", wd)

	_, err := GetResources()
	assert.IsType(t, (*ProviderAuthError)(nil), err)
}

func TestGetResources_DiskImage(t *testing.T) {
	testutils.SkipWithCreds(t)
	// reset
	terraformExec = nil

	t.Setenv("GOOGLE_OAUTH_ACCESS_TOKEN", "")

	wd := path.Join(testutils.RootDir, "test/terraform/gcp_images")
	viper.Set("workdir", wd)

	wantResources := []resources.Resource{
		resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Name:         "diskImage",
				ResourceType: "google_compute_disk",
				Provider:     providers.GCP,
				Region:       "europe-west9",
			},
			Specs: &resources.ComputeResourceSpecs{
				GpuTypes:          nil,
				HddStorage:        decimal.NewFromFloat(10),
				SsdStorage:        decimal.Zero,
				MemoryMb:          0,
				VCPUs:             0,
				CPUType:           "",
				ReplicationFactor: 1,
			},
		},
	}

	resources, err := GetResources()
	if assert.NoError(t, err) {
		assert.Equal(t, len(wantResources), len(resources))
		for i, resource := range resources {
			wantResource := wantResources[i]
			assert.Equal(t, wantResource, resource)
		}
	}

}
