package bigquery

import (
	"context"

	"github.com/odpf/guardian/domain"
)

// Provider for bigquery
type Provider struct {
	typeName string
	clients  map[string]*Client
	crypto   domain.Crypto
}

// NewProvider returns bigquery provider
func NewProvider(typeName string, crypto domain.Crypto) *Provider {
	return &Provider{
		typeName: typeName,
		clients:  map[string]*Client{},
		crypto:   crypto,
	}
}

// GetType returns the provider type
func (p *Provider) GetType() string {
	return p.typeName
}

// CreateConfig validates provider config
func (p *Provider) CreateConfig(pc *domain.ProviderConfig) error {
	c := NewConfig(pc, p.crypto)

	if err := c.ParseAndValidate(); err != nil {
		return err
	}

	return c.EncryptCredentials()
}

// GetResources returns BigQuery dataset and table resources
func (p *Provider) GetResources(pc *domain.ProviderConfig) ([]*domain.Resource, error) {
	client, err := p.getClient(pc.URN, Credentials(pc.Credentials.(string)))
	if err != nil {
		return nil, err
	}

	resources := []*domain.Resource{}
	ctx := context.Background()
	datasets, err := client.GetDatasets(ctx)
	if err != nil {
		return nil, err
	}
	for _, d := range datasets {
		dataset := d.toDomain()
		dataset.ProviderType = pc.Type
		dataset.ProviderURN = pc.URN
		resources = append(resources, dataset)

		tables, err := client.GetTables(ctx, dataset.Name)
		if err != nil {
			return nil, err
		}
		for _, t := range tables {
			table := t.toDomain()
			table.ProviderType = pc.Type
			table.ProviderURN = pc.URN
			resources = append(resources, table)
		}
	}

	return resources, nil
}

func (p *Provider) getClient(projectID string, credentials Credentials) (*Client, error) {
	if p.clients[projectID] != nil {
		return p.clients[projectID], nil
	}

	credentials.Decrypt(p.crypto)
	client, err := NewClient(projectID, []byte(credentials))
	if err != nil {
		return nil, err
	}

	p.clients[projectID] = client
	return client, nil
}
