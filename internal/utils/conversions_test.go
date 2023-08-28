package utils

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseToInt(t *testing.T) {
	type args struct {
		value interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{"int", args{1}, 1, false},
		{"float", args{1.0}, 1, false},
		{"string", args{"1"}, 1, false},
		{"stringFloat", args{"1.0"}, 1, false},
		{"stringErr", args{"a"}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseToInt(tt.args.value)
			assert.Equal(t, got, tt.want)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConvertInterfaceListToStringList(t *testing.T) {
	type args struct {
		list []interface{}
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"empty", args{[]interface{}{}}, []string{}},
		{"one", args{[]interface{}{"a"}}, []string{"a"}},
		{"two", args{[]interface{}{"a", "b"}}, []string{"a", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertInterfaceListToStringList(tt.args.list)
			assert.True(t, reflect.DeepEqual(got, tt.want))
		})
	}
}
