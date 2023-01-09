# Adding new provider

### Introduce new provider type

```go title="domain/provider.go"
package domain

...

const (
    ...
	// ProviderTypeNoOp is the type name for No-Op provider
	ProviderTypeNoOp = "noop"
)
```

### Initialize the provider

```go title="internal/server/services.go"
import (
    ...
	"github.com/odpf/guardian/plugins/providers/noop"
)

...

func InitServices(deps ServiceDeps) (*Services, error) {
    ...
	providerClients := []provider.Client{
		...
		noop.NewProvider(domain.ProviderTypeNoOp, deps.Logger),
	}
```

### Provider implementation

#### Interfaces

Provider should implement `provider.Client`, `providers.PermissionManager` and `providers.Client` interface

```go title="core/provider/service.go"
type Client interface {
	providers.PermissionManager
	providers.Client
}
```

```go title="plugins/providers/client.go"
type Client interface {
	GetType() string
	CreateConfig(*domain.ProviderConfig) error
	GetResources(pc *domain.ProviderConfig) ([]*domain.Resource, error)
	GrantAccess(*domain.ProviderConfig, domain.Grant) error
	RevokeAccess(*domain.ProviderConfig, domain.Grant) error
	GetRoles(pc *domain.ProviderConfig, resourceType string) ([]*domain.Role, error)
	GetAccountTypes() []string
	ListAccess(context.Context, domain.ProviderConfig, []*domain.Resource) (domain.MapResourceAccess, error)
}

type PermissionManager interface {
	GetPermissions(p *domain.ProviderConfig, resourceType, role string) ([]interface{}, error)
}
```

#### Example NoOp Provider

```go title="plugins/providers/noop/provider.go"
package noop

...

type Provider struct {
	provider.UnimplementedClient
	provider.PermissionManager

	typeName string

	logger log.Logger
}

func NewProvider(typeName string, logger log.Logger) *Provider {
	return &Provider{
		typeName: typeName,

		logger: logger,
	}
}

func (p *Provider) GetType() string {
	return p.typeName
}

func (p *Provider) CreateConfig(cfg *domain.ProviderConfig) error {
	// CreateConfig implementation
}

func (p *Provider) GetResources(pc *domain.ProviderConfig) ([]*domain.Resource, error) {
	// GetResources implementation
}

func (p *Provider) GrantAccess(*domain.ProviderConfig, domain.Grant) error {
	// GrantAccess implementation
}

func (p *Provider) RevokeAccess(*domain.ProviderConfig, domain.Grant) error {
	// RevokeAccess implementation
}

func (p *Provider) GetRoles(pc *domain.ProviderConfig, resourceType string) ([]*domain.Role, error) {
	// GetRoles implementation
}

func (p *Provider) GetAccountTypes() []string {
	// GetAccountTypes implementation
}
```

See full implementation here
- [bigquery](https://github.com/odpf/guardian/tree/main/plugins/providers/bigquery)
- [gcloudiam](https://github.com/odpf/guardian/tree/main/plugins/providers/gcloudiam)
- [gcs](https://github.com/odpf/guardian/tree/main/plugins/providers/gcs)
- [grafana](https://github.com/odpf/guardian/tree/main/plugins/providers/grafana)
- [metabase](https://github.com/odpf/guardian/tree/main/plugins/providers/metabase)
- [noop](https://github.com/odpf/guardian/tree/main/plugins/providers/noop)
- [tableau](https://github.com/odpf/guardian/tree/main/plugins/providers/tableau)