package estimate

import (
	"github.com/carboniferio/carbonifer/pkg/providers"
	"github.com/carboniferio/carbonifer/pkg/resources"
	"github.com/shopspring/decimal"
	"reflect"
	"testing"
)

func TestGetEstimation(t *testing.T) {
	type args struct {
		resource resources.GenericResource
	}
	tests := []struct {
		name    string
		args    args
		want    EstimationReport
		wantErr bool
	}{
		{
			name: "e2-standard-2",
			args: args{
				resource: resources.GenericResource{
					Name:     "e2-standard-2",
					Region:   "europe-west4",
					Provider: providers.GCP,
					CPUTypes: []string{
						"Skylake",
						"Broadwell",
						"Haswell",
						"AMD EPYC Rome",
						"AMD EPYC Milan",
					},
					VCPUs:             2,
					MemoryMb:          8192,
					Storage:           resources.Storage{},
					ReplicationFactor: 0,
				},
			},
			want: EstimationReport{
				Resource: resources.GenericResource{
					Name:     "e2-standard-2",
					Region:   "europe-west4",
					Provider: providers.GCP,
					CPUTypes: []string{
						"Skylake",
						"Broadwell",
						"Haswell",
						"AMD EPYC Rome",
						"AMD EPYC Milan",
					},
					MemoryMb:          8192,
					VCPUs:             2,
					Storage:           resources.Storage{},
					ReplicationFactor: 0,
				},
				Power:           decimal.NewFromFloat(4.8677297799),
				CarbonEmissions: decimal.NewFromFloat(1377.5675277218),
				AverageCPUUsage: decimal.NewFromInt(0),
				Count:           decimal.NewFromInt(1),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetEstimation(tt.args.resource)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEstimation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEstimation() got = %v, want %v", got, tt.want)
			}
		})
	}
}
