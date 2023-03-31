package newpoc_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/goto/guardian/domain"
	"github.com/goto/guardian/plugins/providers/newpoc"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/iam/v1"
)

func TestClient_GetAllowedAccountTypes(t *testing.T) {
	type fields struct {
		providerConfig              *domain.ProviderConfig
		cloudResourceManagerService *cloudresourcemanager.Service
		iamService                  *iam.Service
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			name: "should return allowed account types",
			fields: fields{
				providerConfig:              &domain.ProviderConfig{},
				cloudResourceManagerService: &cloudresourcemanager.Service{},
				iamService:                  &iam.Service{},
			},
			args: args{
				ctx: context.Background(),
			},
			want: []string{
				newpoc.AccountTypeUser,
				newpoc.AccountTypeServiceAccount,
				newpoc.AccountTypeGroup,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := newpoc.NewClient(
				&newpoc.ClientDependencies{
					ProviderConfig:              tt.fields.providerConfig,
					CloudResourceManagerService: tt.fields.cloudResourceManagerService,
					IamService:                  tt.fields.iamService,
				},
			)
			if err != nil {
				t.Errorf("NewClient() error = %v", err)
				return
			}

			if got := c.GetAllowedAccountTypes(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.GetAllowedAccountTypes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_ListResources(t *testing.T) {
	type fields struct {
		providerConfig              *domain.ProviderConfig
		cloudResourceManagerService *cloudresourcemanager.Service
		iamService                  *iam.Service
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []newpoc.IResource
		wantErr bool
	}{
		{
			name: "should return list of resources with type project",
			fields: fields{
				providerConfig: &domain.ProviderConfig{
					Type: "newpoc",
					URN:  "newpoc",
					Credentials: map[string]interface{}{
						"service_account_key": "service_account_key",
						"resource_name":       "projects/test",
					},
				},
				cloudResourceManagerService: &cloudresourcemanager.Service{},
				iamService:                  &iam.Service{},
			},
			args: args{
				ctx: context.Background(),
			},
			want: []newpoc.IResource{
				&domain.Resource{
					ProviderType: "newpoc",
					ProviderURN:  "newpoc",
					Type:         newpoc.ResourceTypeProject,
					URN:          "projects/test",
					Name:         fmt.Sprintf("%s - GCP IAM", "projects/test"),
				},
			},
		},
		{
			name: "should return list of resources with type organization",
			fields: fields{
				providerConfig: &domain.ProviderConfig{
					Type: "newpoc",
					URN:  "newpoc",
					Credentials: map[string]interface{}{
						"service_account_key": "service_account_key",
						"resource_name":       "organizations/test",
					},
				},
				cloudResourceManagerService: &cloudresourcemanager.Service{},
				iamService:                  &iam.Service{},
			},
			args: args{
				ctx: context.Background(),
			},
			want: []newpoc.IResource{
				&domain.Resource{
					ProviderType: "newpoc",
					ProviderURN:  "newpoc",
					Type:         newpoc.ResourceTypeOrganization,
					URN:          "organizations/test",
					Name:         fmt.Sprintf("%s - GCP IAM", "organizations/test"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := newpoc.NewClient(
				&newpoc.ClientDependencies{
					ProviderConfig:              tt.fields.providerConfig,
					CloudResourceManagerService: tt.fields.cloudResourceManagerService,
					IamService:                  tt.fields.iamService,
				},
			)
			if err != nil {
				t.Errorf("NewClient() error = %v", err)
				return
			}

			got, err := c.ListResources(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.ListResources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("Client.ListResources() has %v elements, want %v elements", len(got), len(tt.want))
			}
			for i := range got {
				if !reflect.DeepEqual(got[i], tt.want[i]) {
					t.Errorf("Client.ListResources()[%v] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
