package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapToSlice(t *testing.T) {
	type args struct {
		m map[string]string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty map",
			args: args{
				m: map[string]string{},
			},
			want: []string{},
		},
		{
			name: "map with one key/value pair",
			args: args{
				m: map[string]string{
					"key1": "value1",
				},
			},
			want: []string{"key1=value1"},
		},
		{
			name: "map with multiple key/value pairs",
			args: args{
				m: map[string]string{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				},
			},
			want: []string{"key1=value1", "key2=value2", "key3=value3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.ElementsMatchf(t, tt.want, MapToSlice(tt.args.m), "MapToSlice(%v)", tt.args.m)
		})
	}
}
