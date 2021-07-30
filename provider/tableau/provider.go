package tableau

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
	workbooks, err := client.GetWorkbooks()
	if err != nil {
		return nil, err
	}
	for _, w := range workbooks {
		wb := w.ToDomain()
		wb.ProviderType = pc.Type
		wb.ProviderURN = pc.URN
		resources = append(resources, wb)
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
	client, err := newClient(&ClientConfig{
		Host:       credentials.Host,
		Username:   credentials.Username,
		Password:   credentials.Password,
		ContentURL: credentials.ContentURL,
	})
	if err != nil {
		return nil, err
	}

	p.clients[providerURN] = client
	return client, nil
}
