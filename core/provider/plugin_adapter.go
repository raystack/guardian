package provider

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/goto/guardian/domain"
	"github.com/goto/guardian/plugins/providers/newpoc"
)

type pluginAdapter struct {
	providerType        string
	allowedAccountTypes []string
	factory             *pluginFactory

	validator *validator.Validate
	crypto    domain.Crypto
}

func (a *pluginAdapter) GetType() string {
	return a.providerType
}

func (a *pluginAdapter) CreateConfig(pc *domain.ProviderConfig) error {
	config, err := a.factory.getConfig(pc)
	if err != nil {
		return fmt.Errorf("initializing config for %q: %w", pc.Type, err)
	}

	if err := config.Validate(context.TODO(), a.validator); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	if encryptableConfig, ok := config.(encryptable); ok {
		if err := encryptableConfig.Encrypt(a.crypto); err != nil {
			return fmt.Errorf("encrypting config: %w", err)
		}
	}

	return nil
}

func (a *pluginAdapter) GetResources(pc *domain.ProviderConfig) ([]*domain.Resource, error) {
	config, err := a.factory.getConfig(pc)
	if err != nil {
		return nil, fmt.Errorf("initializing config for %q: %w", pc.URN, err)
	}
	client, err := a.factory.getClient(config)
	if err != nil {
		return nil, fmt.Errorf("initializing client for %q: %w", pc.URN, err)
	}

	resourceables, err := client.ListResources(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("listing resources for %q: %w", pc.URN, err)
	}

	resources := make([]*domain.Resource, 0, len(resourceables))
	for _, resourceable := range resourceables {
		r := &domain.Resource{
			ProviderType: pc.Type,
			ProviderURN:  pc.URN,
			Type:         resourceable.GetType(),
			URN:          resourceable.GetURN(),
			Name:         resourceable.GetDisplayName(),
		}
		if md := resourceable.GetMetadata(); md != nil {
			r.Details = map[string]interface{}{
				"__metadata": resourceable.GetMetadata(),
			}
		}
		resources = append(resources, r)
	}

	return resources, nil
}

func (a *pluginAdapter) GrantAccess(pc *domain.ProviderConfig, grant domain.Grant) error {
	config, err := a.factory.getConfig(pc)
	if err != nil {
		return fmt.Errorf("initializing config for %q: %w", pc.URN, err)
	}
	client, err := a.factory.getClient(config)
	if err != nil {
		return fmt.Errorf("initializing client for %q: %w", pc.URN, err)
	}

	return client.GrantAccess(context.TODO(), grant)
}

func (a *pluginAdapter) RevokeAccess(pc *domain.ProviderConfig, grant domain.Grant) error {
	config, err := a.factory.getConfig(pc)
	if err != nil {
		return fmt.Errorf("initializing config for %q: %w", pc.URN, err)
	}
	client, err := a.factory.getClient(config)
	if err != nil {
		return fmt.Errorf("initializing client for %q: %w", pc.URN, err)
	}

	return client.RevokeAccess(context.TODO(), grant)
}

func (a *pluginAdapter) GetRoles(pc *domain.ProviderConfig, resourceType string) ([]*domain.Role, error) {
	for _, r := range pc.Resources {
		if r.Type == resourceType {
			return r.Roles, nil
		}
	}

	return nil, ErrInvalidResourceType
}

func (a *pluginAdapter) GetAccountTypes() []string {
	return a.allowedAccountTypes
}

func (a *pluginAdapter) ListAccess(ctx context.Context, pc domain.ProviderConfig, resources []*domain.Resource) (domain.MapResourceAccess, error) {
	config, err := a.factory.getConfig(&pc)
	if err != nil {
		return nil, fmt.Errorf("initializing config for %q: %w", pc.URN, err)
	}
	client, err := a.factory.getClient(config)
	if err != nil {
		return nil, fmt.Errorf("initializing client for %q: %w", pc.URN, err)
	}

	if accessImporter, ok := client.(AccessImporter); ok {
		return accessImporter.ListAccess(ctx, pc, resources)
	}

	return nil, fmt.Errorf("ListAccess %w", ErrUnimplementedMethod)
}

func (a *pluginAdapter) GetPermissions(pc *domain.ProviderConfig, resourceType, role string) ([]interface{}, error) {
	for _, rc := range pc.Resources {
		if rc.Type != resourceType {
			continue
		}
		for _, r := range rc.Roles {
			if r.ID == role {
				if r.Permissions == nil {
					return make([]interface{}, 0), nil
				}
				return r.Permissions, nil
			}
		}
		return nil, ErrInvalidRole
	}
	return nil, ErrInvalidResourceType
}

func getNewPlugins(pluginFactory *pluginFactory, validator *validator.Validate, crypto domain.Crypto) map[string]Client {
	return map[string]Client{
		newpoc.ProviderType: &pluginAdapter{
			providerType: newpoc.ProviderType,
			allowedAccountTypes: []string{
				newpoc.AccountTypeUser,
				newpoc.AccountTypeGroup,
				newpoc.AccountTypeServiceAccount,
			},
			factory:   pluginFactory,
			validator: validator,
			crypto:    crypto,
		},
	}
}
