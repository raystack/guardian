package newpoc

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/goto/guardian/domain"
)

// domain/provider.go
//
// type ProviderConfigEncryptor interface {
// 	Encrypt(context.Context, *Provider) error
// }
//
// type Provider struct{}
// func (p *Provider) Encrypt(ctx context.Context, e ProviderConfigEncryptor) error {
// 	return e.Encrypt(ctx, p)
// }

// plugin will export two main structs. 1. ConfigManager, 2. Client

// TODO: all of these interfaces should be defined in core/provider/service.go only

type ProviderConfigurator interface {
	Validate(*validator.Validate) error
	Encrypt(domain.Encryptor) error
	Decrypt(domain.Decryptor) error
}

// BasicProviderClient depends on a valid provider config
type BasicProviderClient interface {
	GetAllowedAccountTypes(context.Context) []string
	ListResources(context.Context) ([]IResource, error)
	GrantAccess(context.Context, domain.Grant) error
	RevokeAccess(context.Context, domain.Grant) error
}

type IResource interface {
	GetType() string
	GetURN() string
	GetDisplayName() string
	GetMetadata() map[string]interface{}
}

type AccessImporter interface {
	ListAccess(context.Context, domain.ProviderConfig) (domain.MapResourceAccess, error)
}

type ActivityImporter interface {
	ListActivities(context.Context) ([]IActivity, error)
}

type IActivity interface {
	GetID() string
	// TODO: complete methods of activity interface
}

// type Dataset struct {}
// func (d Dataset) GetType() string { return "dataset"}

// type Table struct {}
// func (t Table) GetType() string { return "table"}

// type IBigQueryResource interface {
// 	Dataset | Table
// }

// type BigQueryResource[T IBigQueryResource] struct {

// }

// in provider service struct:
// cached bigqueryClient for provider A (credentials A)
// cached bigqueryClient for provider B (credentials B)
// cached gcsCLient for provider C (credentials C)

// provider config:
//   resource type config:
//     roles config:
//     - id: my-custom-role-1
//       permissions: roleA, roleB
//     - id: my-custom-role-2
//       permissions: roleB, roleC

// gcp
// role: roles/viewer, roles/bigquery.dataViewer, etc.
// permissions: projects.list, projects.get, datasets.list, etc.

// pv.PermissionManager
// grafana
// metabase
// shield
// tableau

// provider.PermissionManager
// bigquery
// gcloudiam
// gcs
// noop
