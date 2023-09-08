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
	_ "github.com/carboniferio/carbonifer/internal/testutils"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var persistentDisk tfjson.StateResource = tfjson.StateResource{
	Address:      "google_compute_disk.disk1",
	Type:         "google_compute_disk",
	Name:         "disk1",
	ProviderName: "google",
	AttributeValues: map[string]interface{}{
		"name": "disk1",
		"type": "pd-standard",
		"size": float64(1024),
		"zone": "europe-west9-a",
	},
}

var persistentDiskNoSize tfjson.StateResource = tfjson.StateResource{
	Address:      "google_compute_disk.disk2",
	Type:         "google_compute_disk",
	Name:         "disk2",
	ProviderName: "google",
	AttributeValues: map[string]interface{}{
		"name": "disk2",
		"type": "pd-standard",
		"zone": "europe-west9-a",
	},
}

var regionDisk tfjson.StateResource = tfjson.StateResource{
	Address:      "google_compute_region_disk.diskr",
	Type:         "google_compute_region_disk",
	Name:         "diskr",
	ProviderName: "google",
	AttributeValues: map[string]interface{}{
		"name":          "diskr",
		"type":          "pd-ssd",
		"size":          float64(1024),
		"replica_zones": []interface{}{"europe-west9-a", "europe-west9-b"},
	},
}

var gpuAttachedMachine tfjson.StateResource = tfjson.StateResource{
	Address:      "google_compute_instance.attachedgpu",
	Type:         "google_compute_instance",
	Name:         "attachedgpu",
	ProviderName: "google",
	AttributeValues: map[string]interface{}{
		"name":         "attachedgpu",
		"machine_type": "n1-standard-2",
		"zone":         "europe-west9-a",
		"boot_disk":    []interface{}{},
		"guest_accelerator": []interface{}{
			map[string]interface{}{
				"type":  "nvidia-tesla-k80",
				"count": float64(2),
			},
		},
	},
}

var gpuDefaultMachine tfjson.StateResource = tfjson.StateResource{
	Address:      "google_compute_instance.defaultgpu",
	Type:         "google_compute_instance",
	Name:         "defaultgpu",
	ProviderName: "google",
	AttributeValues: map[string]interface{}{
		"name":         "defaultgpu",
		"machine_type": "a2-highgpu-1g",
		"zone":         "europe-west9-a",
		"boot_disk":    []interface{}{},
	},
}

func TestGetResource(t *testing.T) {
	mapping, err := plan.GetMapping()
	assert.NoError(t, err)
	computeResourceMapping := *mapping.ComputeResource
	type args struct {
		tfResource tfjson.StateResource
		mapping    plan.ResourceMapping
	}
	tests := []struct {
		name string
		args args
		want resources.Resource
	}{
		{
			name: "diskWithSize",
			args: args{
				tfResource: persistentDisk,
				mapping:    computeResourceMapping["google_compute_disk"],
			},
			want: resources.ComputeResource{
				Identification: &resources.ResourceIdentification{
					Address:      "google_compute_disk.disk1",
					Name:         "disk1",
					ResourceType: "google_compute_disk",
					Provider:     providers.GCP,
					Region:       "europe-west9",
					Count:        1,
				},
				Specs: &resources.ComputeResourceSpecs{
					HddStorage:        decimal.New(1024, 0),
					SsdStorage:        decimal.Zero,
					ReplicationFactor: 1,
				},
			},
		},
		{
			name: "diskWithNoSize",
			args: args{
				tfResource: persistentDiskNoSize,
				mapping:    computeResourceMapping["google_compute_disk"],
			},
			want: resources.ComputeResource{
				Identification: &resources.ResourceIdentification{
					Address:      "google_compute_disk.disk2",
					Name:         "disk2",
					ResourceType: "google_compute_disk",
					Provider:     providers.GCP,
					Region:       "europe-west9",
					Count:        1,
				},
				Specs: &resources.ComputeResourceSpecs{
					HddStorage:        decimal.New(10, 0),
					SsdStorage:        decimal.Zero,
					ReplicationFactor: 1,
				},
			},
		},
		{
			name: "regionDisk",
			args: args{
				tfResource: regionDisk,
				mapping:    computeResourceMapping["google_compute_disk"],
			},
			want: resources.ComputeResource{
				Identification: &resources.ResourceIdentification{
					Address:      "google_compute_region_disk.diskr",
					Name:         "diskr",
					ResourceType: "google_compute_region_disk",
					Provider:     providers.GCP,
					Region:       "europe-west9",
					Count:        1,
				},
				Specs: &resources.ComputeResourceSpecs{
					HddStorage:        decimal.Zero,
					SsdStorage:        decimal.NewFromInt(1024),
					ReplicationFactor: 2,
				},
			},
		},
		{
			name: "gpu attached",
			args: args{
				tfResource: gpuAttachedMachine,
				mapping:    computeResourceMapping["google_compute_instance"],
			},
			want: resources.ComputeResource{
				Identification: &resources.ResourceIdentification{
					Address:      "google_compute_instance.attachedgpu",
					Name:         "attachedgpu",
					ResourceType: "google_compute_instance",
					Provider:     providers.GCP,
					Region:       "europe-west9",
					Count:        1,
				},
				Specs: &resources.ComputeResourceSpecs{
					VCPUs:    int32(2),
					MemoryMb: int32(7680),
					GpuTypes: []string{
						"nvidia-tesla-k80",
						"nvidia-tesla-k80",
					},
					HddStorage:        decimal.Zero,
					SsdStorage:        decimal.Zero,
					ReplicationFactor: 1,
				},
			},
		},
		{
			name: "gpu default",
			args: args{
				tfResource: gpuDefaultMachine,
				mapping:    computeResourceMapping["google_compute_instance"],
			},
			want: resources.ComputeResource{
				Identification: &resources.ResourceIdentification{
					Address:      "google_compute_instance.defaultgpu",
					Name:         "defaultgpu",
					ResourceType: "google_compute_instance",
					Provider:     providers.GCP,
					Region:       "europe-west9",
					Count:        1,
				},
				Specs: &resources.ComputeResourceSpecs{
					GpuTypes:          nil,
					VCPUs:             int32(12),
					MemoryMb:          int32(87040),
					HddStorage:        decimal.Zero,
					SsdStorage:        decimal.Zero,
					ReplicationFactor: 1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource, _ := testutils.TfResourceToJSON(&tt.args.tfResource)
			got, err := plan.GetComputeResource(*resource, &tt.args.mapping, nil)
			assert.NoError(t, err)
			assert.Len(t, got, 1)
			assert.IsType(t, resources.ComputeResource{}, got[0])
			gotResource := got[0].(resources.ComputeResource)
			assert.Equal(t, tt.want, gotResource)
			assert.NoError(t, err)
		})
	}
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
				Address:      "google_compute_disk.diskImage",
				Name:         "diskImage",
				ResourceType: "google_compute_disk",
				Provider:     providers.GCP,
				Region:       "europe-west9",
				Count:        1,
			},
			Specs: &resources.ComputeResourceSpecs{
				GpuTypes:          nil,
				HddStorage:        decimal.New(int64(10), 0),
				SsdStorage:        decimal.Zero,
				MemoryMb:          0,
				VCPUs:             0,
				CPUType:           "",
				ReplicationFactor: 1,
			},
		},
	}

	tfPlan, _ := terraform.TerraformPlan()
	resourceList, err := plan.GetResources(tfPlan)
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

	wd := path.Join(testutils.RootDir, "test/terraform/gcp_group")
	viper.Set("workdir", wd)

	wantResources := map[string]resources.Resource{
		"google_compute_network.vpc_network": resources.UnsupportedResource{
			Identification: &resources.ResourceIdentification{
				Address:      "google_compute_network.vpc_network",
				Name:         "vpc_network",
				ResourceType: "google_compute_network",
				Provider:     providers.GCP,
				Region:       "",
				Count:        1,
			},
		},
		"google_compute_subnetwork.first": resources.UnsupportedResource{
			Identification: &resources.ResourceIdentification{
				Address:      "google_compute_subnetwork.first",
				Name:         "first",
				ResourceType: "google_compute_subnetwork",
				Provider:     providers.GCP,
				Region:       "",
				Count:        1,
			},
		},
		"google_compute_instance_group_manager.my-group-manager": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Address:      "google_compute_instance_group_manager.my-group-manager",
				Name:         "my-group-manager",
				ResourceType: "google_compute_instance_group_manager",
				Provider:     providers.GCP,
				Region:       "europe-west9",
				Count:        3,
			},
			Specs: &resources.ComputeResourceSpecs{
				GpuTypes:          nil,
				HddStorage:        decimal.New(20, 0),
				SsdStorage:        decimal.Zero,
				MemoryMb:          8192,
				VCPUs:             2,
				CPUType:           "",
				ReplicationFactor: 1,
			},
		},
	}

	tfPlan, _ := terraform.TerraformPlan()
	resources, err := plan.GetResources(tfPlan)
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

	wd := path.Join(testutils.RootDir, "test/terraform/gcp_cit")
	viper.Set("workdir", wd)

	wantResources := map[string]resources.Resource{
		"google_compute_network.vpc_network": resources.UnsupportedResource{
			Identification: &resources.ResourceIdentification{
				Address:      "google_compute_network.vpc_network",
				Name:         "vpc_network",
				ResourceType: "google_compute_network",
				Provider:     providers.GCP,
				Region:       "",
				Count:        1,
			},
		},
		"google_compute_subnetwork.first": resources.UnsupportedResource{
			Identification: &resources.ResourceIdentification{
				Address:      "google_compute_subnetwork.first",
				Name:         "first",
				ResourceType: "google_compute_subnetwork",
				Provider:     providers.GCP,
				Region:       "",
				Count:        1,
			},
		},
		"google_compute_instance_from_template.ifromtpl": resources.ComputeResource{
			Identification: &resources.ResourceIdentification{
				Address:      "google_compute_instance_from_template.ifromtpl",
				Name:         "ifromtpl",
				ResourceType: "google_compute_instance_from_template",
				Provider:     providers.GCP,
				Region:       "europe-west9",
				Count:        1,
			},
			Specs: &resources.ComputeResourceSpecs{
				GpuTypes:          nil,
				HddStorage:        decimal.New(20, 0),
				SsdStorage:        decimal.Zero,
				MemoryMb:          8192,
				VCPUs:             2,
				CPUType:           "",
				ReplicationFactor: 1,
			},
		},
	}

	tfPlan, _ := terraform.TerraformPlan()
	resources, err := plan.GetResources(tfPlan)
	if assert.NoError(t, err) {
		for i, resource := range resources {
			wantResource := wantResources[i]
			assert.EqualValues(t, wantResource, resource)
		}
	}

}
