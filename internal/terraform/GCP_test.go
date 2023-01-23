package terraform

import (
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

func TestGetResource(t *testing.T) {
	type args struct {
		tfResource tfjson.StateResource
	}
	tests := []struct {
		name string
		args args
		want resources.Resource
	}{
		{
			name: "diskWithSize",
			args: args{
				tfResource: persistenDisk,
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
				tfResource: persistenDiskNoSize,
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
				tfResource: regionDisk,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetResource(tt.args.tfResource, nil)
			assert.Equal(t, tt.want, got)
		})
	}
}
