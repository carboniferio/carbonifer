package gcp

import (
	"testing"

	_ "github.com/carboniferio/carbonifer/internal/testutils"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestGetGCPMachineType(t *testing.T) {
	type args struct {
		machineTypeStr string
		zone           string
	}
	tests := []struct {
		name string
		args args
		want MachineType
	}{
		{
			name: "existing",
			args: args{"e2-standard-2", "europe-west9-a"},
			want: MachineType{
				Name:     "e2-standard-2",
				Vcpus:    2,
				GPUTypes: nil,
				MemoryMb: 8192,
				CpuTypes: []string{
					"Skylake", "Broadwell", "Haswell", "AMD EPYC Rome", "AMD EPYC Milan",
				},
			},
		},
		// {
		// 	name: "custom",
		// 	args: args{"custom-2-2048", "europe-west9-a"},
		// 	want: MachineType{
		// 		Name:     "custom-2-2048",
		// 		Vcpus:    2,
		// 		GPUTypes: nil,
		// 		MemoryMb: 2048,
		// 	},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetGCPMachineType(tt.args.machineTypeStr, tt.args.zone)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestGetCPUWatt(t *testing.T) {
	got := GetCPUWatt("Skylake")
	want := CPUWatt{
		Architecture:        "Skylake",
		MinWatts:            decimal.NewFromFloat(0.6446044454253452),
		MaxWatts:            decimal.NewFromFloat(3.8984738056304855),
		GridCarbonIntensity: decimal.NewFromFloat(80.43037974683544),
	}
	assert.Equal(t, got, want)
}
