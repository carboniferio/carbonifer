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
	logrus.SetLevel(logrus.DebugLevel)
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
	logrus.SetLevel(logrus.DebugLevel)

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
	logrus.SetLevel(logrus.DebugLevel)

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
	logrus.SetLevel(logrus.DebugLevel)

	wd := path.Join(testutils.RootDir, "test/terraform/gcp_1")
	viper.Set("workdir", wd)

	wantResources := []resources.Resource{
		resources.ComputeResource{
			Identification: &resources.ComputeResourceIdentification{
				Name:         "first",
				ResourceType: "google_compute_disk",
				Provider:     providers.GCP,
				Region:       "europe-west9",
			},
			Specs: &resources.ComputeResourceSpecs{
				Gpu:               0,
				HddStorage:        decimal.NewFromInt(1024),
				SsdStorage:        decimal.Zero,
				MemoryMb:          0,
				VCPUs:             0,
				CPUType:           "",
				ReplicationFactor: 1,
			},
		},
		resources.ComputeResource{
			Identification: &resources.ComputeResourceIdentification{
				Name:         "first",
				ResourceType: "google_compute_instance",
				Provider:     providers.GCP,
				Region:       "europe-west9",
			},
			Specs: &resources.ComputeResourceSpecs{
				Gpu:        0,
				HddStorage: decimal.Zero,
				SsdStorage: decimal.NewFromFloat(567).Add(decimal.NewFromFloat(375).Add(decimal.NewFromFloat(375))),
				MemoryMb:   2480,
				VCPUs:      1,
				CPUType:    "",
			},
		},
		resources.ComputeResource{
			Identification: &resources.ComputeResourceIdentification{
				Name:         "second",
				ResourceType: "google_compute_instance",
				Provider:     providers.GCP,
				Region:       "europe-west9",
			},
			Specs: &resources.ComputeResourceSpecs{
				Gpu:        0,
				HddStorage: decimal.NewFromFloat(10),
				SsdStorage: decimal.Zero,
				MemoryMb:   4098,
				VCPUs:      2,
				CPUType:    "",
			},
		},
		resources.UnsupportedResource{
			Identification: &resources.ComputeResourceIdentification{
				Name:         "vpc_network",
				ResourceType: "google_compute_network",
				Provider:     providers.GCP,
				Region:       "",
			},
		},
		resources.ComputeResource{
			Identification: &resources.ComputeResourceIdentification{
				Name:         "regional-first",
				ResourceType: "google_compute_region_disk",
				Provider:     providers.GCP,
				Region:       "europe-west9",
			},
			Specs: &resources.ComputeResourceSpecs{
				Gpu:               0,
				HddStorage:        decimal.NewFromInt(1024),
				SsdStorage:        decimal.Zero,
				MemoryMb:          0,
				VCPUs:             0,
				CPUType:           "",
				ReplicationFactor: 2,
			},
		},
		resources.UnsupportedResource{
			Identification: &resources.ComputeResourceIdentification{
				Name:         "first",
				ResourceType: "google_compute_subnetwork",
				Provider:     providers.GCP,
				Region:       "europe-west9",
			},
		},
	}

	resources := GetResources()
	assert.Equal(t, len(resources), len(wantResources))
	for i, resource := range resources {
		wantResource := wantResources[i]
		assert.Equal(t, wantResource, resource)
	}
}
