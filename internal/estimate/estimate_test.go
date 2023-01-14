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
	Name:         "machine-name-1",
	ResourceType: "type-1",
	Provider:     providers.GCP,
	Region:       "europe-west9",
	VCPUs:        2,
	MemoryMb:     4096,
}

var resourceGCPComputeCPUType = resources.ComputeResource{
	Name:         "machine-name-2",
	ResourceType: "type-1",
	Provider:     providers.GCP,
	Region:       "europe-west9",
	VCPUs:        2,
	MemoryMb:     4096,
	CPUType:      "Broadwell",
}

var resourceAWSComputeBasic = resources.ComputeResource{
	Name:         "machine-name-1",
	ResourceType: "type-1",
	Provider:     providers.AWS,
	Region:       "europe-west9",
	VCPUs:        2,
	MemoryMb:     4096,
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
				Resource:        resourceGCPComputeBasic,
				Power:           decimal.NewFromFloat(7.600784000).Round(10),
				CarbonEmissions: decimal.NewFromFloat(0.448446256).Round(10),
				AverageCPUUsage: decimal.NewFromFloat(avg_cpu_use),
			},
		},
		{
			name: "gcp_specific_cpu_type",
			args: args{resourceGCPComputeCPUType},
			want: &EstimationResource{
				Resource:        resourceGCPComputeCPUType,
				Power:           decimal.NewFromFloat(6.5760185741),
				CarbonEmissions: decimal.NewFromFloat(0.3879850958),
				AverageCPUUsage: decimal.NewFromFloat(avg_cpu_use),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := EstimateResource(tt.args.resource)
			EqualsEstimationResource(t, got, tt.want)
		})
	}
}

func TestEstimateResourceKilo(t *testing.T) {
	avg_cpu_use := viper.GetFloat64("avg_cpu_use")
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
				Resource:        resourceGCPComputeBasic,
				Power:           decimal.NewFromFloat(7.600784000 * 24 * 30).RoundFloor(10),
				CarbonEmissions: decimal.NewFromFloat(232.4745391104).RoundFloor(10),
				AverageCPUUsage: decimal.NewFromFloat(avg_cpu_use),
			},
		},
		{
			name: "gcp_specific_cpu_type",
			args: args{resourceGCPComputeCPUType},
			want: &EstimationResource{
				Resource:        resourceGCPComputeCPUType,
				Power:           decimal.NewFromFloat(4734.7333733647).RoundFloor(10),
				CarbonEmissions: decimal.NewFromFloat(201.1314737005).RoundFloor(10),
				AverageCPUUsage: decimal.NewFromFloat(avg_cpu_use),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := EstimateResource(tt.args.resource)
			//assert.Equal(t, got.Power, tt.want.Power)
			EqualsEstimationResource(t, got, tt.want)
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

func EqualsEstimationResource(t *testing.T, res1 *EstimationResource, res2 *EstimationResource) {
	assert.Equal(t, res1.Resource, res2.Resource)
	assert.Equal(t, res1.Power.String(), res2.Power.String())
	assert.Equal(t, res1.CarbonEmissions.String(), res2.CarbonEmissions.String())
	assert.Equal(t, res1.AverageCPUUsage.String(), res2.AverageCPUUsage.String())
	// return reflect.DeepEqual(res1.Resource, res2.Resource) &&
	// 	res1.Power.String() == res2.Power.String() &&
	// 	res1.CarbonEmissions.String() == res2.CarbonEmissions.String() &&
	// 	res1.AverageCPUUsage.String() == res2.AverageCPUUsage.String()
}

func EqualsTotal(t *testing.T, res1 *EstimationTotal, res2 *EstimationTotal) {
	assert.Equal(t, res1.ResourcesCount, res2.ResourcesCount)
	assert.Equal(t, res1.Power.String(), res2.Power.String())
	assert.Equal(t, res1.CarbonEmissions.String(), res2.CarbonEmissions.String())
}

func TestEstimateResources(t *testing.T) {
	avg_cpu_use := viper.GetFloat64("avg_cpu_use")
	viper.Set("unit.carbon", "g")
	viper.Set("unit.time", "h")
	type args struct {
		resources []resources.ComputeResource
	}
	tests := []struct {
		name string
		args args
		want EstimationReport
	}{
		{
			name: "gcp_array",
			args: args{
				[]resources.ComputeResource{
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
						Resource:        resourceGCPComputeBasic,
						Power:           decimal.NewFromFloat(7.600784000).Round(10),
						CarbonEmissions: decimal.NewFromFloat(0.448446256).Round(10),
						AverageCPUUsage: decimal.NewFromFloat(avg_cpu_use),
					},
					{
						Resource:        resourceGCPComputeCPUType,
						Power:           decimal.NewFromFloat(6.5760185741),
						CarbonEmissions: decimal.NewFromFloat(0.3879850958),
						AverageCPUUsage: decimal.NewFromFloat(avg_cpu_use),
					},
				},
				Total: EstimationTotal{
					Power:           decimal.NewFromFloat(0.8364313518),
					CarbonEmissions: decimal.NewFromFloat(0.8364313518),
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
				EqualsEstimationResource(t, &gotResource, &wantResource)
			}

			EqualsTotal(t, &got.Total, &tt.want.Total)
		})
	}
}
