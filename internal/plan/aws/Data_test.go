package aws

import (
	"testing"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
)

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
			name: "AMI with ebs 20 Gb",
			args: args{
				tfResource: tfjson.StateResource{
					Address: "data.aws_ami.foo",
					Type:    "aws_ami",
					Name:    "foo",
					AttributeValues: map[string]interface{}{
						"name": "foo",
						"block_device_mappings": []interface{}{
							map[string]interface{}{
								"device_name": "/dev/sda1",
								"ebs": map[string]interface{}{
									"volume_size": "20",
									"volume_type": "gp2",
								},
							},
						},
						"id": "ami-1234567890",
					},
				},
			},
			want: resources.EbsDataResource{
				Identification: &resources.ResourceIdentification{
					Name:         "foo",
					ResourceType: "aws_ami",
					Provider:     providers.AWS,
				},
				DataImageSpecs: []*resources.DataImageSpecs{
					{
						DiskSizeGb: 20,
						DeviceName: "/dev/sda1",
						VolumeType: "gp2",
					},
				},
				AwsId: "ami-1234567890",
			},
		},
		{
			name: "Snapshot of size 60 Gb",
			args: args{
				tfResource: tfjson.StateResource{
					Address: "data.aws_ebs_snapshot.test_snapshot",
					Type:    "aws_ebs_snapshot",
					Name:    "test_snapshot",
					AttributeValues: map[string]interface{}{
						"id":          "snap-1234567890",
						"volume_size": float64(60),
					},
				},
			},
			want: resources.EbsDataResource{
				Identification: &resources.ResourceIdentification{
					Name:         "test_snapshot",
					ResourceType: "aws_ebs_snapshot",
					Provider:     providers.AWS,
				},
				DataImageSpecs: []*resources.DataImageSpecs{
					{
						DiskSizeGb: 60,
					},
				},
				AwsId: "snap-1234567890",
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
