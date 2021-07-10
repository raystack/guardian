package grafana

import (
	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/domain"
)

type provider struct {
	typeName string
	clients  map[string]*client
	crypto   domain.Crypto
}

func NewProvider(typeName string, crypto domain.Crypto) *provider {
	return &provider{
		typeName: typeName,
		clients:  map[string]*client{},
		crypto:   crypto,
	}
}

func (p *provider) GetType() string {
	return p.typeName
}

func (p *provider) CreateConfig(pc *domain.ProviderConfig) error {
	c := NewConfig(pc, p.crypto)

	if err := c.ParseAndValidate(); err != nil {
		return err
	}

	return c.EncryptCredentials()
}

func (p *provider) GetResources(pc *domain.ProviderConfig) ([]*domain.Resource, error) {
	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return nil, err
	}

	client, err := p.getClient(pc.URN, creds)
	if err != nil {
		return nil, err
	}

	resources := []*domain.Resource{}

	folders, err := client.getFolders()
	if err != nil {
		return nil, err
	}
	for _, f := range folders {
		fd := f.toDomain()
		dashboards, err := client.getDashboards(fd.ID)
		if err != nil {
			return nil, err
		}
		for _, d := range dashboards {
			db := d.toDomain()
			db.ProviderType = pc.Type
			db.ProviderURN = pc.URN
			resources = append(resources, db)
		}
	}
	return resources, nil
}

func (p *provider) GrantAccess(pc *domain.ProviderConfig, a *domain.Appeal) error {
	//TO DO
	return nil
}

func (p *provider) RevokeAccess(pc *domain.ProviderConfig, a *domain.Appeal) error {
	//TO DO
	return nil
}

func (p *provider) getClient(providerURN string, credentials Credentials) (*client, error) {
	if p.clients[providerURN] != nil {
		return p.clients[providerURN], nil
	}

	credentials.Decrypt(p.crypto)
	client, err := NewClient(&ClientConfig{
		Host:   credentials.Host,
		ApiKey: credentials.ApiKey,
	})
	if err != nil {
		return nil, err
	}

	p.clients[providerURN] = client
	return client, nil
}
