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

func TestIsValidUUID(t *testing.T) {
	type args struct {
		u string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid uuid",
			args: args{
				u: "123e4567-e89b-12d3-a456-426614174000",
			},
			want: true,
		},
		{
			name: "invalid uuid",
			args: args{
				u: "123e4567-a456-42661417400",
			},
			want: false,
		},
		{
			name: "empty uuid",
			args: args{
				u: "",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, IsValidUUID(tt.args.u), "IsValidUUID(%v)", tt.args.u)
		})
	}
}

func TestIsInteger(t *testing.T) {
	type args struct {
		val float64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "is integer",
			args: args{
				val: 1,
			},
			want: true,
		},
		{
			name: "is not integer",
			args: args{
				val: 1.1,
			},
			want: false,
		},
		{
			name: "is not integer",
			args: args{
				val: 0.1,
			},
			want: false,
		},
		{
			name: "is integer",
			args: args{
				val: 0.0,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, IsInteger(tt.args.val), "IsInteger(%v)", tt.args.val)
		})
	}
}
