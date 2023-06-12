package gcp

import (
	"testing"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	_ "github.com/carboniferio/carbonifer/internal/testutils"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
)

var imageResource tfjson.StateResource = tfjson.StateResource{
	Address: "data.google_compute_image.debian",
	Mode:    "data",
	Type:    "google_compute_image",
	Name:    "debian",
	AttributeValues: map[string]interface{}{
		"name":         "debian-11-bullseye-v20221206",
		"self_link":    "https://www.googleapis.com/compute/v1/projects/debian-cloud/global/images/debian-11-bullseye-v20221206",
		"disk_size_gb": float64(10),
	},
}

func TestGetDataResource(t *testing.T) {
	type args struct {
		tfResource tfjson.StateResource
	}
	tests := []struct {
		name string
		args args
		want resources.DataResource
	}{
		{
			name: "existing",
			args: args{
				tfResource: imageResource,
			},
			want: resources.DataImageResource{
				Identification: &resources.ResourceIdentification{
					Name:         "debian",
					ResourceType: "google_compute_image",
					Provider:     providers.GCP},
				DataImageSpecs: []*resources.DataImageSpecs{
					{
						DiskSizeGb: float64(10),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDataResource(tt.args.tfResource)
			assert.Equal(t, tt.want, got)
		})
	}
}
