package aws

import (
	"testing"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/carboniferio/carbonifer/internal/terraform/tfrefs"
	_ "github.com/carboniferio/carbonifer/internal/testutils"
	tfjson "github.com/hashicorp/terraform-json"
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
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetResource(tt.args.tfResource, tt.args.tfRefs)
			assert.Equal(t, tt.want, got)
		})
	}
}
