package estimate

import (
	"reflect"
	"testing"

	"github.com/carboniferio/carbonifer/internal/providers"
	"github.com/carboniferio/carbonifer/internal/resources"
	_ "github.com/carboniferio/carbonifer/internal/testutils"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var resourceGCPComputeBasic = resources.ComputeResource{
	Identification: &resources.ResourceIdentification{
		Name:         "machine-name-1",
		ResourceType: "type-1",
		Provider:     providers.GCP,
		Region:       "europe-west9",
	},
	Specs: &resources.ComputeResourceSpecs{
		VCPUs:    2,
		MemoryMb: 4096,
	},
}

var resourceGCPComputeCPUType = resources.ComputeResource{
	Identification: &resources.ResourceIdentification{
		Name:         "machine-name-2",
		ResourceType: "type-1",
		Provider:     providers.GCP,
		Region:       "europe-west9",
	},
	Specs: &resources.ComputeResourceSpecs{
		VCPUs:      2,
		MemoryMb:   4096,
		CPUType:    "Broadwell",
		SsdStorage: decimal.NewFromFloat(1024),
		HddStorage: decimal.NewFromFloat(2044),
		GpuTypes: []string{
			"nvidia-t4",
			"nvidia-tesla-a100",
		},
	},
}

var resourceAWSComputeBasic = resources.ComputeResource{
	Identification: &resources.ResourceIdentification{
		Name:         "machine-name-1",
		ResourceType: "type-1",
		Provider:     providers.AWS,
		Region:       "europe-west9",
	},
	Specs: &resources.ComputeResourceSpecs{
		VCPUs:    2,
		MemoryMb: 4096,
	},
}

func TestEstimateResource(t *testing.T) {
	avg_cpu_use := viper.GetFloat64("avg_cpu_use")
	type args struct {
		resource resources.ComputeResource
	}
	tests := []struct {
		name string
		args args
		want *EstimationResource
	}{
		{
			name: "gcp_basic",
			args: args{resourceGCPComputeBasic},
			want: &EstimationResource{
				Resource:        &resourceGCPComputeBasic,
				Power:           decimal.NewFromFloat(7.600784000).RoundFloor(10),
				CarbonEmissions: decimal.NewFromFloat(0.448446256).RoundFloor(10),
				AverageCPUUsage: decimal.NewFromFloat(avg_cpu_use),
			},
		},
		{
			name: "gcp_specific_cpu_type",
			args: args{resourceGCPComputeCPUType},
			want: &EstimationResource{
				Resource:        &resourceGCPComputeCPUType,
				Power:           decimal.NewFromFloat(318.1165660741),
				CarbonEmissions: decimal.NewFromFloat(18.7688773983),
				AverageCPUUsage: decimal.NewFromFloat(avg_cpu_use),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := EstimateResource(tt.args.resource)
			EqualsEstimationResource(t, tt.want, got)
		})
	}
}

func TestEstimateResourceKilo(t *testing.T) {
	avg_cpu_use := viper.GetFloat64("provider.gcp.avg_cpu_use")
	viper.Set("unit.carbon", "kg")
	viper.Set("unit.time", "m")
	type args struct {
		resource resources.ComputeResource
	}
	tests := []struct {
		name string
		args args
		want *EstimationResource
	}{
		{
			name: "gcp_basic",
			args: args{resourceGCPComputeBasic},
			want: &EstimationResource{
				Resource:        &resourceGCPComputeBasic,
				Power:           decimal.NewFromFloat(5472.56448).RoundFloor(10),
				CarbonEmissions: decimal.NewFromFloat(232.4745391104).RoundFloor(10),
				AverageCPUUsage: decimal.NewFromFloat(avg_cpu_use),
			},
		},
		{
			name: "gcp_specific_cpu_type",
			args: args{resourceGCPComputeCPUType},
			want: &EstimationResource{
				Resource:        &resourceGCPComputeCPUType,
				Power:           decimal.NewFromFloat(229043.9275733647).RoundFloor(10),
				CarbonEmissions: decimal.NewFromFloat(9729.7860433165).RoundFloor(10),
				AverageCPUUsage: decimal.NewFromFloat(avg_cpu_use),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := EstimateResource(tt.args.resource)
			EqualsEstimationResource(t, tt.want, got)
		})
	}
}

func TestEstimateResourceUnsupported(t *testing.T) {
	type args struct {
		resource resources.ComputeResource
	}
	tests := []struct {
		name string
		args args
		want *providers.UnsupportedProviderError
	}{
		{
			name: "gcp_basic",
			args: args{resourceAWSComputeBasic},
			want: &providers.UnsupportedProviderError{Provider: "AWS"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EstimateResource(tt.args.resource)
			//assert.Equal(t, got.Power, tt.want.Power)
			if !reflect.DeepEqual(err, tt.want) {
				t.Errorf("EstimateResource() = %v, want %v", err, tt.want)
			}
		})
	}
}

func EqualsEstimationResource(t *testing.T, expected *EstimationResource, actual *EstimationResource) {
	assert.Equal(t, expected.Resource, actual.Resource)
	assert.Equal(t, expected.Power.String(), actual.Power.String())
	assert.Equal(t, expected.CarbonEmissions.String(), actual.CarbonEmissions.String())
	assert.Equal(t, expected.AverageCPUUsage.String(), actual.AverageCPUUsage.String())

}

func EqualsTotal(t *testing.T, expected *EstimationTotal, actual *EstimationTotal) {
	assert.Equal(t, expected.ResourcesCount, actual.ResourcesCount)
	assert.Equal(t, expected.Power.String(), actual.Power.String())
	assert.Equal(t, expected.CarbonEmissions.String(), actual.CarbonEmissions.String())
}

func TestEstimateResources(t *testing.T) {
	avg_cpu_use := viper.GetFloat64("provider.gcp.avg_cpu_use")
	viper.Set("unit.carbon", "g")
	viper.Set("unit.time", "h")
	type args struct {
		resources []resources.Resource
	}
	tests := []struct {
		name string
		args args
		want EstimationReport
	}{
		{
			name: "gcp_array",
			args: args{
				[]resources.Resource{
					resourceGCPComputeBasic,
					resourceGCPComputeCPUType,
				},
			},
			want: EstimationReport{
				Info: EstimationInfo{
					UnitTime:                "h",
					UnitWattTime:            "Wh",
					UnitCarbonEmissionsTime: "gCO2eq/h",
				},
				Resources: []EstimationResource{
					{
						Resource:        &resourceGCPComputeBasic,
						Power:           decimal.NewFromFloat(7.600784).Round(10),
						CarbonEmissions: decimal.NewFromFloat(0.448446256).Round(10),
						AverageCPUUsage: decimal.NewFromFloat(avg_cpu_use),
					},
					{
						Resource:        &resourceGCPComputeCPUType,
						Power:           decimal.NewFromFloat(318.1165660741),
						CarbonEmissions: decimal.NewFromFloat(18.7688773983),
						AverageCPUUsage: decimal.NewFromFloat(avg_cpu_use),
					},
				},
				Total: EstimationTotal{
					Power:           decimal.NewFromFloat(325.7173500741),
					CarbonEmissions: decimal.NewFromFloat(19.2173236543),
					ResourcesCount:  2,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateResources(tt.args.resources)
			assert.Equal(t, got.Info.UnitCarbonEmissionsTime, tt.want.Info.UnitCarbonEmissionsTime)
			assert.Equal(t, got.Info.UnitTime, tt.want.Info.UnitTime)
			assert.Equal(t, got.Info.UnitWattTime, tt.want.Info.UnitWattTime)
			for i, gotResource := range got.Resources {
				wantResource := tt.want.Resources[i]
				EqualsEstimationResource(t, &wantResource, &gotResource)
			}

			EqualsTotal(t, &tt.want.Total, &got.Total)
		})
	}
}
