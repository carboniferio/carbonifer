package aws

import (
	"reflect"
	"testing"

	_ "github.com/carboniferio/carbonifer/internal/testutils"
)

func TestGetAWSInstanceType(t *testing.T) {
	type args struct {
		instanceTypeStr string
	}
	tests := []struct {
		name string
		args args
		want InstanceType
	}{
		{
			name: "c5d.12xlarge",
			args: args{instanceTypeStr: "c5d.12xlarge"},
			want: InstanceType{
				InstanceType: "c5d.12xlarge",
				VCPU:         48,
				MemoryMb:     96 * 1024,
				InstanceStorage: InstanceStorage{
					SizePerDiskGB: 900,
					Count:         2,
					Type:          "ssd",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetAWSInstanceType(tt.args.instanceTypeStr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAWSInstanceType() = %v, want %v", got, tt.want)
			}
		})
	}
}
