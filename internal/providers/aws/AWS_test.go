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
		want MachineType
	}{
		{
			name: "t2.micro",
			args: args{instanceTypeStr: "t2.micro"},
			want: MachineType{
				InstanceType: "t2.micro",
				VCPU:         1,
				MemoryMb:     1024,
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
