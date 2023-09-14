package notifiers

import (
	"github.com/goto/guardian/plugins/notifiers/slack"
	"reflect"
	"testing"
)

func TestNewSlackConfig(t *testing.T) {
	type args struct {
		config *Config
	}
	tests := []struct {
		name    string
		args    args
		want    *slack.Config
		wantErr bool
	}{
		{
			name: "should return error when no access token or workspaces are provided",
			args: args{
				config: &Config{
					Provider: ProviderTypeSlack,
				},
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "should return error when both access token and workspaces are provided",
			args: args{
				config: &Config{
					Provider:    ProviderTypeSlack,
					AccessToken: "foo",
					SlackConfig: SlackConfig{
						"workspaces": []slack.SlackWorkspace{
							{
								WorkspaceName: "default",
								AccessToken:   "bar",
								Criteria:      "1==1",
							},
						},
					},
				},
			},
			want:    nil,
			wantErr: true,
		}, {
			name: "should return slack config when access token is provided",
			args: args{
				config: &Config{
					Provider:    ProviderTypeSlack,
					AccessToken: "foo",
				},
			},
			want: &slack.Config{
				Workspaces: []slack.SlackWorkspace{
					{
						WorkspaceName: "default",
						AccessToken:   "foo",
						Criteria:      "1==1",
					},
				},
			},
			wantErr: false,
		}, {
			name: "should return slack config when workspaces are provided",
			args: args{
				config: &Config{
					Provider: ProviderTypeSlack,
					SlackConfig: SlackConfig{
						"workspaces": []slack.SlackWorkspace{
							{
								WorkspaceName: "A",
								AccessToken:   "foo",
								Criteria:      "$email contains '@abc'",
							},
							{
								WorkspaceName: "B",
								AccessToken:   "bar",
								Criteria:      "$email contains '@xyz'",
							},
						},
					},
				},
			},
			want: &slack.Config{
				Workspaces: []slack.SlackWorkspace{
					{
						WorkspaceName: "A",
						AccessToken:   "foo",
						Criteria:      "$email contains '@abc'",
					},
					{
						WorkspaceName: "B",
						AccessToken:   "bar",
						Criteria:      "$email contains '@xyz'",
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSlackConfig(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSlackConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSlackConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
