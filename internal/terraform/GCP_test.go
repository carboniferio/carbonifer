package terraform

import (
	"reflect"
	"strings"
	"testing"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

var persistenDisk tfjson.StateResource = tfjson.StateResource{
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

var persistenDiskNoSize tfjson.StateResource = tfjson.StateResource{
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
		"machine_type": "n2-standard-2",
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
	type args struct {
		tfResource tfjson.ConfigResource
	}
	tests := []struct {
		name string
		args args
		want resources.Resource
	}{
		{
			name: "diskWithSize",
			args: args{
				tfResource: StateResourceToConfigResource(persistenDisk),
			},
			want: resources.ComputeResource{
				Identification: &resources.ResourceIdentification{
					Name:         "disk1",
					ResourceType: "google_compute_disk",
					Provider:     providers.GCP,
					Region:       "europe-west9",
				},
				Specs: &resources.ComputeResourceSpecs{
					HddStorage:        decimal.NewFromInt(1024),
					SsdStorage:        decimal.Zero,
					ReplicationFactor: 1,
				},
			},
		},
		{
			name: "diskWithNoSize",
			args: args{
				tfResource: StateResourceToConfigResource(persistenDiskNoSize),
			},
			want: resources.ComputeResource{
				Identification: &resources.ResourceIdentification{
					Name:         "disk2",
					ResourceType: "google_compute_disk",
					Provider:     providers.GCP,
					Region:       "europe-west9",
				},
				Specs: &resources.ComputeResourceSpecs{
					HddStorage:        decimal.New(50, 1),
					SsdStorage:        decimal.Zero,
					ReplicationFactor: 1,
				},
			},
		},
		{
			name: "regionDisk",
			args: args{
				tfResource: StateResourceToConfigResource(regionDisk),
			},
			want: resources.ComputeResource{
				Identification: &resources.ResourceIdentification{
					Name:         "diskr",
					ResourceType: "google_compute_region_disk",
					Provider:     providers.GCP,
					Region:       "europe-west9",
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
				tfResource: StateResourceToConfigResource(gpuAttachedMachine),
			},
			want: resources.ComputeResource{
				Identification: &resources.ResourceIdentification{
					Name:         "attachedgpu",
					ResourceType: "google_compute_instance",
					Provider:     providers.GCP,
					Region:       "europe-west9",
				},
				Specs: &resources.ComputeResourceSpecs{
					GpuTypes: []string{
						"nvidia-tesla-k80",
						"nvidia-tesla-k80",
					},
					HddStorage: decimal.Zero,
					SsdStorage: decimal.Zero,
				},
			},
		},
		{
			name: "gpu default",
			args: args{
				tfResource: StateResourceToConfigResource(gpuDefaultMachine),
			},
			want: resources.ComputeResource{
				Identification: &resources.ResourceIdentification{
					Name:         "defaultgpu",
					ResourceType: "google_compute_instance",
					Provider:     providers.GCP,
					Region:       "europe-west9",
				},
				Specs: &resources.ComputeResourceSpecs{
					GpuTypes: []string{
						"nvidia-tesla-a100",
					},
					VCPUs:      int32(12),
					MemoryMb:   int32(87040),
					HddStorage: decimal.Zero,
					SsdStorage: decimal.Zero,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetResource(tt.args.tfResource, nil)
			assert.Equal(t, tt.want, got)
		})
	}
}

func StateResourceToConfigResource(stateResource tfjson.StateResource) tfjson.ConfigResource {
	mode := tfjson.ManagedResourceMode
	if strings.HasPrefix(stateResource.Address, "data.") {
		mode = tfjson.DataResourceMode
	}
	expressions := make(map[string]*tfjson.Expression)
	for key, value := range stateResource.AttributeValues {
		expressionData := tfjson.ExpressionData{
			ConstantValue: nil,
			References:    nil,
			NestedBlocks:  nil,
		}

		rt := reflect.TypeOf(value)
		if rt.Kind() == reflect.Slice || rt.Kind() == reflect.Array {
			len := len(value.([]interface{}))
			if len > 0 {
				expressionData.ConstantValue = value
			}
		} else {
			expressionData.ConstantValue = value
		}
		expr := tfjson.Expression{
			ExpressionData: &expressionData,
		}
		expressions[key] = &expr
	}
	return tfjson.ConfigResource{
		Address:           stateResource.Address,
		Mode:              mode,
		Type:              stateResource.Type,
		Name:              stateResource.Name,
		ProviderConfigKey: strings.Split(stateResource.Address, "_")[0],
		Expressions:       expressions,
		SchemaVersion:     stateResource.SchemaVersion,
	}
}
