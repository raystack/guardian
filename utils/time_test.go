package utils

import "testing"

func TestGetReadableDuration(t *testing.T) {
	type args struct {
		durationStr string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "should return duration in integral days when input can be converted into days - 1",
			args: args{
				durationStr: "2160h",
			},
			want:    "90d",
			wantErr: false,
		},
		{
			name: "should return duration in integral days when input can be converted into days - 2",
			args: args{
				durationStr: "24h",
			},
			want:    "1d",
			wantErr: false,
		},
		{
			name: "should return duration in integral days when input can be converted into days - 3",
			args: args{
				durationStr: "86400s",
			},
			want:    "1d",
			wantErr: false,
		},
		{
			name: "should return duration same as input when input cannot be converted into days - 1",
			args: args{
				durationStr: "1h",
			},
			want:    "1h",
			wantErr: false,
		},
		{
			name: "should return duration same as input when input cannot be converted into days - 2",
			args: args{
				durationStr: "25h",
			},
			want:    "25h",
			wantErr: false,
		},
		{
			name: "should return duration same as input when input cannot be converted into days - 3",
			args: args{
				durationStr: "86399s",
			},
			want:    "86399s",
			wantErr: false,
		},
		{
			name: "should return duration same as input when input is not a valid duration",
			args: args{
				durationStr: "124hs",
			},
			want:    "124hs",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetReadableDuration(tt.args.durationStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetReadableDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetReadableDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
