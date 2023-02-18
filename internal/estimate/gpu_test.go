package estimate

import (
	"testing"

	"github.com/carboniferio/carbonifer/internal/resources"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

var noGPUResource = resources.ComputeResource{
	Identification: &resources.ResourceIdentification{
		Name:  "no-gpu",
		Count: 1,
	},
	Specs: &resources.ComputeResourceSpecs{
		GpuTypes: nil,
	},
}

var twoGPUResources = resources.ComputeResource{
	Identification: &resources.ResourceIdentification{
		Name:  "two-gpu",
		Count: 1,
	},
	Specs: &resources.ComputeResourceSpecs{
		GpuTypes: []string{
			"nvidia-t4",
			"nvidia-tesla-a100",
		},
	},
}

func Test_estimateWattGPU(t *testing.T) {
	type args struct {
		resource *resources.ComputeResource
	}
	tests := []struct {
		name string
		args args
		want decimal.Decimal
	}{
		{
			name: "NO GPU",
			args: args{&noGPUResource},
			want: decimal.Zero,
		},
		{
			name: "Two Default GPU",
			args: args{&twoGPUResources},
			want: decimal.New(2660, -1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := estimateWattGPU(tt.args.resource)
			assert.Equal(t, tt.want, got)

		})
	}
}
