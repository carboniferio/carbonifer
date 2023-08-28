package terraform

import (
	"fmt"
	"testing"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/carboniferio/carbonifer/internal/testutils"
	_ "github.com/carboniferio/carbonifer/internal/testutils"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

var persistentDisk tfjson.StateResource = tfjson.StateResource{
	Address: "google_compute_disk.disk1",
	Type:    "google_compute_disk",
	Name:    "disk1",
	AttributeValues: map[string]interface{}{
		"name": "disk1",
		"type": "pd-standard",
		"size": float64(1024),
		"zone": "europe-west9-a",
	},
}

var persistentDiskNoSize tfjson.StateResource = tfjson.StateResource{
	Address: "google_compute_disk.disk2",
	Type:    "google_compute_disk",
	Name:    "disk2",
	AttributeValues: map[string]interface{}{
		"name": "disk2",
		"type": "pd-standard",
		"zone": "europe-west9-a",
	},
}

var regionDisk tfjson.StateResource = tfjson.StateResource{
	Address: "google_compute_region_disk.diskr",
	Type:    "google_compute_region_disk",
	Name:    "diskr",
	AttributeValues: map[string]interface{}{
		"name":          "diskr",
		"type":          "pd-ssd",
		"size":          float64(1024),
		"replica_zones": []interface{}{"europe-west9-a", "europe-west9-b"},
	},
}

var gpuAttachedMachine tfjson.StateResource = tfjson.StateResource{
	Address: "google_compute_instance.attachedgpu",
	Type:    "google_compute_instance",
	Name:    "attachedgpu",
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
	Address: "google_compute_instance.defaultgpu",
	Type:    "google_compute_instance",
	Name:    "defaultgpu",
	AttributeValues: map[string]interface{}{
		"name":         "defaultgpu",
		"machine_type": "a2-highgpu-1g",
		"zone":         "europe-west9-a",
		"boot_disk":    []interface{}{},
	},
}

func TestGetResource(t *testing.T) {
	mapping, err := getMapping(providers.GCP)
	assert.NoError(t, err)
	computeResourceMapping := *mapping.computeResource
	type args struct {
		tfResource tfjson.StateResource
		mapping    map[string]interface{}
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
			got, err := getComputeResource(*resource, tt.args.mapping, nil)
			assert.Len(t, got, 1)
			assert.IsType(t, resources.ComputeResource{}, got[0])
			gotResource := got[0].(resources.ComputeResource)
			fmt.Println(gotResource.Specs.SsdStorage)
			assert.Equal(t, tt.want, gotResource)
			assert.NoError(t, err)
		})
	}
}
