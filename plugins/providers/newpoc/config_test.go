package newpoc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/goto/guardian/domain"
	"github.com/goto/guardian/plugins/providers/gcs/mocks"
	"github.com/goto/guardian/plugins/providers/newpoc"
	"github.com/stretchr/testify/mock"
)

func TestConfigManager_Validate(t *testing.T) {
	type fields struct {
		validator *validator.Validate
		crypto    domain.Crypto
	}
	type args struct {
		ctx context.Context
		p   *domain.Provider
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "should return error if provider is nil",
			fields: fields{
				validator: validator.New(),
				crypto:    nil,
			},
			args: args{
				ctx: context.Background(),
				p:   nil,
			},
			wantErr: true,
		},
		{
			name: "should return error decoding credentials",
			fields: fields{
				validator: validator.New(),
				crypto:    nil,
			},
			args: args{
				ctx: context.Background(),
				p: &domain.Provider{
					Config: &domain.ProviderConfig{
						Credentials: "invalid",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "should return error if credentials are invalid",
			fields: fields{
				validator: validator.New(),
				crypto:    nil,
			},
			args: args{
				ctx: context.Background(),
				p: &domain.Provider{
					Config: &domain.ProviderConfig{
						Credentials: map[string]interface{}{
							"service_account_key": "",
							"resource_name":       "",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "should return error if resource length is not 1",
			fields: fields{
				validator: validator.New(),
				crypto:    nil,
			},
			args: args{
				ctx: context.Background(),
				p: &domain.Provider{
					Config: &domain.ProviderConfig{
						Credentials: map[string]interface{}{
							"service_account_key": "service_account_key",
							"resource_name":       "projects/test",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "should return error if resource name is invalid",
			fields: fields{
				validator: validator.New(),
				crypto:    nil,
			},
			args: args{
				ctx: context.Background(),
				p: &domain.Provider{
					Config: &domain.ProviderConfig{
						Credentials: map[string]interface{}{
							"service_account_key": "service_account_key",
							"resource_name":       "projects/test",
						},
						Resources: []*domain.ResourceConfig{
							{
								Type: "invalid",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "should return error if roles is empty",
			fields: fields{
				validator: validator.New(),
				crypto:    nil,
			},
			args: args{
				ctx: context.Background(),
				p: &domain.Provider{
					Config: &domain.ProviderConfig{
						Credentials: map[string]interface{}{
							"service_account_key": "service_account_key",
							"resource_name":       "projects/test",
						},
						Resources: []*domain.ResourceConfig{
							{
								Type: "project",
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "should return nil if config is valid",
			fields: fields{
				validator: validator.New(),
				crypto:    nil,
			},
			args: args{
				ctx: context.Background(),
				p: &domain.Provider{
					Config: &domain.ProviderConfig{
						Credentials: map[string]interface{}{
							"service_account_key": "service_account_key",
							"resource_name":       "projects/test",
						},
						Resources: []*domain.ResourceConfig{
							{
								Type: "project",
								Roles: []*domain.Role{
									{
										Name: "roles/owner",
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newpoc.NewConfigManager(
				tt.fields.validator,
				tt.fields.crypto,
			)
			if err := m.Validate(tt.args.ctx, tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("ConfigManager.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigManager_Encrypt(t *testing.T) {
	type fields struct {
		validator *validator.Validate
		crypto    domain.Crypto
	}
	type args struct {
		ctx context.Context
		p   *domain.Provider
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "should return error decoding credentials",
			fields: fields{
				validator: validator.New(),
				crypto:    nil,
			},
			args: args{
				ctx: context.Background(),
				p: &domain.Provider{
					Config: &domain.ProviderConfig{
						Credentials: "invalid",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "should return error if encrypt fails",
			fields: fields{
				validator: validator.New(),
				crypto: func() domain.Crypto {
					c := new(mocks.Crypto)
					c.On("Encrypt", mock.Anything, mock.Anything).Return("", errors.New("error"))
					return c
				}(),
			},
			args: args{
				ctx: context.Background(),
				p: &domain.Provider{
					Config: &domain.ProviderConfig{
						Credentials: map[string]interface{}{
							"service_account_key": "service_account_key",
							"resource_name":       "projects/test",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "should return nil if encrypt succeeds",
			fields: fields{
				validator: validator.New(),
				crypto: func() domain.Crypto {
					c := new(mocks.Crypto)
					c.On("Encrypt", mock.Anything, mock.Anything).Return("encrypted", nil)
					return c
				}(),
			},
			args: args{
				ctx: context.Background(),
				p: &domain.Provider{
					Config: &domain.ProviderConfig{
						Credentials: map[string]interface{}{
							"service_account_key": "service_account_key",
							"resource_name":       "projects/test",
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newpoc.NewConfigManager(
				tt.fields.validator,
				tt.fields.crypto,
			)
			if err := m.Encrypt(tt.args.ctx, tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("ConfigManager.Encrypt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigManager_Decrypt(t *testing.T) {
	type fields struct {
		validator *validator.Validate
		crypto    domain.Crypto
	}
	type args struct {
		ctx context.Context
		p   *domain.Provider
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "should return error decoding credentials",
			fields: fields{
				validator: validator.New(),
				crypto:    nil,
			},
			args: args{
				ctx: context.Background(),
				p: &domain.Provider{
					Config: &domain.ProviderConfig{
						Credentials: "invalid",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "should return error if decrypt fails",
			fields: fields{
				validator: validator.New(),
				crypto: func() domain.Crypto {
					c := new(mocks.Crypto)
					c.On("Decrypt", mock.Anything, mock.Anything).Return("", errors.New("error"))
					return c
				}(),
			},
			args: args{
				ctx: context.Background(),
				p: &domain.Provider{
					Config: &domain.ProviderConfig{
						Credentials: map[string]interface{}{
							"service_account_key": "service_account_key",
							"resource_name":       "projects/test",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "should return nil if decrypt succeeds",
			fields: fields{
				validator: validator.New(),
				crypto: func() domain.Crypto {
					c := new(mocks.Crypto)
					c.On("Decrypt", mock.Anything, mock.Anything).Return("encrypted", nil)
					return c
				}(),
			},
			args: args{
				ctx: context.Background(),
				p: &domain.Provider{
					Config: &domain.ProviderConfig{
						Credentials: map[string]interface{}{
							"service_account_key": "service_account_key",
							"resource_name":       "projects/test",
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newpoc.NewConfigManager(
				tt.fields.validator,
				tt.fields.crypto,
			)
			if err := m.Decrypt(tt.args.ctx, tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("ConfigManager.Decrypt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
