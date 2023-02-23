package resources

import (
	"github.com/carboniferio/carbonifer/pkg/providers"
	"reflect"
	"testing"
)

func TestGenericResource_IsSupported(t *testing.T) {
	type fields struct {
		Name     string
		Region   string
		Provider providers.Provider
		GPUTypes []string
		CPUTypes []string
		MemoryMb int32
		Storage  Storage
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "GCP",
			fields: fields{
				Provider: providers.GCP,
			},
			want: true,
		},
		{
			name: "AWS",
			fields: fields{
				Provider: providers.AWS,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := GenericResource{
				Name:     tt.fields.Name,
				Region:   tt.fields.Region,
				Provider: tt.fields.Provider,
				GPUTypes: tt.fields.GPUTypes,
				CPUTypes: tt.fields.CPUTypes,
				MemoryMb: tt.fields.MemoryMb,
				Storage:  tt.fields.Storage,
			}
			if got := g.IsSupported(); got != tt.want {
				t.Errorf("IsSupported() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetResource(t *testing.T) {
	type args struct {
		instanceType string
		zone         string
		provider     providers.Provider
	}
	tests := []struct {
		name    string
		args    args
		want    GenericResource
		wantErr bool
	}{
		{
			name: "e2-standard-2",
			args: args{
				instanceType: "e2-standard-2",
				zone:         "europe-west4-a",
				provider:     providers.GCP,
			},
			want: GenericResource{
				Name:     "e2-standard-2",
				Region:   "europe-west4-a",
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
				Storage:           Storage{},
				ReplicationFactor: 0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetResource(tt.args.instanceType, tt.args.zone, tt.args.provider)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetResource() got = %v, want %v", got, tt.want)
			}
		})
	}
}
