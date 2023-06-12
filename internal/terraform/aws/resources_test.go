package aws

import (
	"fmt"
	"testing"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/carboniferio/carbonifer/internal/terraform/tfrefs"
	_ "github.com/carboniferio/carbonifer/internal/testutils"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

var defaultMachine tfjson.StateResource = tfjson.StateResource{
	Address: "aws_instance.foo",
	Type:    "aws_instance",
	Name:    "foo",
	AttributeValues: map[string]interface{}{
		"name":          "foo",
		"instance_type": "t2.micro",
	},
}

var machineWithDefaultRootDisk tfjson.StateResource = tfjson.StateResource{
	Address: "aws_instance.foo",
	Type:    "aws_instance",
	Name:    "machineWithDefaultRootDisk",
	AttributeValues: map[string]interface{}{
		"name":          "machineWithDefaultRootDisk",
		"instance_type": "t2.micro",
		"root_block_device": []interface{}{
			map[string]interface{}{
				"delete_on_termination": true,
			},
		},
	},
}

var machineWithRootDiskSize tfjson.StateResource = tfjson.StateResource{
	Address: "aws_instance.foo",
	Type:    "aws_instance",
	Name:    "machineWithRootDiskSize",
	AttributeValues: map[string]interface{}{
		"name":          "machineWithRootDiskSize",
		"instance_type": "t2.micro",
		"root_block_device": []interface{}{
			map[string]interface{}{
				"delete_on_termination": true,
				"volume_size":           float64(20),
			},
		},
	},
}

var machineWithEBSSize tfjson.StateResource = tfjson.StateResource{
	Address: "aws_instance.foo",
	Type:    "aws_instance",
	Name:    "machineWithEBSSize",
	AttributeValues: map[string]interface{}{
		"name":          "machineWithEBSSize",
		"instance_type": "t2.micro",
		"ebs_block_device": []interface{}{
			map[string]interface{}{
				"delete_on_termination": true,
				"volume_size":           float64(50),
				"volume_type":           "st1",
			},
		},
	},
}

var machineWithEBSSizeAndEphemeral tfjson.StateResource = tfjson.StateResource{
	Address: "aws_instance.foo",
	Type:    "aws_instance",
	Name:    "machineWithEBSSizeAndEphemeral",
	AttributeValues: map[string]interface{}{
		"name":          "machineWithEBSSizeAndEphemeral",
		"instance_type": "c5d.12xlarge",
		"ebs_block_device": []interface{}{
			map[string]interface{}{
				"delete_on_termination": true,
				"volume_size":           float64(50),
				"volume_type":           "st1",
			},
		},
		"ephemeral_block_device": []interface{}{
			map[string]interface{}{
				"device_name": "ephemeral0",
			},
			map[string]interface{}{
				"device_name": "ephemeral1",
			},
		},
	},
}

var tfRefs *tfrefs.References = &tfrefs.References{
	ProviderConfigs: map[string]string{
		"region": "eu-west-3",
	},
}

func TestGetResource(t *testing.T) {
	type args struct {
		tfResource tfjson.StateResource
		tfRefs     *tfrefs.References
	}
	tests := []struct {
		name string
		args args
		want resources.Resource
	}{
		{
			name: "aws_instance",
			args: args{
				tfResource: defaultMachine,
				tfRefs:     tfRefs,
			},
			want: resources.ComputeResource{
				Identification: &resources.ResourceIdentification{
					Name:         "foo",
					ResourceType: "aws_instance",
					Provider:     providers.AWS,
					Region:       "eu-west-3",
					Count:        1,
				},
				Specs: &resources.ComputeResourceSpecs{
					VCPUs:             int32(1),
					MemoryMb:          int32(1024),
					ReplicationFactor: 1,
					HddStorage:        decimal.Zero,
					SsdStorage:        decimal.NewFromInt(8),
				},
			},
		},
		{
			name: "aws_instance with default root disk",
			args: args{
				tfResource: machineWithDefaultRootDisk,
				tfRefs:     tfRefs,
			},
			want: resources.ComputeResource{
				Identification: &resources.ResourceIdentification{
					Name:         "machineWithDefaultRootDisk",
					ResourceType: "aws_instance",
					Provider:     providers.AWS,
					Region:       "eu-west-3",
					Count:        1,
				},
				Specs: &resources.ComputeResourceSpecs{
					VCPUs:             int32(1),
					MemoryMb:          int32(1024),
					ReplicationFactor: 1,
					HddStorage:        decimal.Zero,
					SsdStorage:        decimal.NewFromInt(8),
				},
			},
		},
		{
			name: "aws_instance with root disk size",
			args: args{
				tfResource: machineWithRootDiskSize,
				tfRefs:     tfRefs,
			},
			want: resources.ComputeResource{
				Identification: &resources.ResourceIdentification{
					Name:         "machineWithRootDiskSize",
					ResourceType: "aws_instance",
					Provider:     providers.AWS,
					Region:       "eu-west-3",
					Count:        1,
				},
				Specs: &resources.ComputeResourceSpecs{
					VCPUs:             int32(1),
					MemoryMb:          int32(1024),
					ReplicationFactor: 1,
					HddStorage:        decimal.Zero,
					SsdStorage:        decimal.NewFromInt(20),
				},
			},
		},
		{
			name: "aws_instance with ebs hdd disk size",
			args: args{
				tfResource: machineWithEBSSize,
				tfRefs:     tfRefs,
			},
			want: resources.ComputeResource{
				Identification: &resources.ResourceIdentification{
					Name:         "machineWithEBSSize",
					ResourceType: "aws_instance",
					Provider:     providers.AWS,
					Region:       "eu-west-3",
					Count:        1,
				},
				Specs: &resources.ComputeResourceSpecs{
					VCPUs:             int32(1),
					MemoryMb:          int32(1024),
					ReplicationFactor: 1,
					HddStorage:        decimal.NewFromInt(50),
					SsdStorage:        decimal.NewFromInt(8),
				},
			},
		},
		{
			name: "aws_instance with ebs hdd disk size and ephemeral",
			args: args{
				tfResource: machineWithEBSSizeAndEphemeral,
				tfRefs:     tfRefs,
			},
			want: resources.ComputeResource{
				Identification: &resources.ResourceIdentification{
					Name:         "machineWithEBSSizeAndEphemeral",
					ResourceType: "aws_instance",
					Provider:     providers.AWS,
					Region:       "eu-west-3",
					Count:        1,
				},
				Specs: &resources.ComputeResourceSpecs{
					VCPUs:             int32(48),
					MemoryMb:          int32(98304),
					ReplicationFactor: 1,
					HddStorage:        decimal.NewFromInt(50),
					SsdStorage:        decimal.NewFromInt(1808),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetResource(tt.args.tfResource, tt.args.tfRefs)
			fmt.Println("Name", got.(resources.ComputeResource).Identification.Name)
			assert.Equal(t, tt.want, got)
		})
	}
}
