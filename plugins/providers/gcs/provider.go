package gcs

import (
	"context"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/mitchellh/mapstructure"
	"github.com/odpf/guardian/core/provider"
	"github.com/odpf/guardian/domain"
	"github.com/odpf/guardian/utils"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type Provider struct {
	typeName string
	Clients  map[string]GcsClient
	crypto   domain.Crypto
}

func NewProvider(typeName string, crypto domain.Crypto) *Provider {
	return &Provider{
		typeName: typeName,
		Clients:  map[string]GcsClient{},
		crypto:   crypto,
	}
}

func (p *Provider) CreateConfig(pc *domain.ProviderConfig) error {
	c := NewConfig(pc, p.crypto)

	if err := c.ParseAndValidate(); err != nil {
		return err
	}

	return c.EncryptCredentials()
}

func (p *Provider) GetResources(pc *domain.ProviderConfig) ([]*domain.Resource, error) {
	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return nil, err
	}

	if err := creds.Decrypt(p.crypto); err != nil {
		return nil, err
	}

	ctx := context.TODO()
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON([]byte(creds.ServiceAccountKey)))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	resourceTypes := []string{}
	for _, rc := range pc.Resources {
		resourceTypes = append(resourceTypes, rc.Type)
	}

	var resources []*domain.Resource

	projectID := strings.Replace(creds.ResourceName, "projects/", "", 1)

	it := client.Buckets(ctx, projectID)
	for {
		battrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		if utils.ContainsString(resourceTypes, ResourceTypeBucket) {
			b := &Bucket{Name: battrs.Name}
			bucketResource := b.toDomain()
			bucketResource.ProviderType = pc.Type
			bucketResource.ProviderURN = pc.URN
			resources = append(resources, bucketResource)
		}

		if utils.ContainsString(resourceTypes, ResourceTypeObject) {
			objIt := client.Bucket(battrs.Name).Objects(ctx, nil)
			for {
				oattrs, err := objIt.Next()
				if err == iterator.Done {
					break
				}
				if err != nil {
					return nil, err
				}

				o := &Object{Name: oattrs.Name}
				objectResource := o.toDomain()
				objectResource.ProviderType = pc.Type
				objectResource.ProviderURN = pc.URN
				resources = append(resources, objectResource)
			}
		}
	}

	return resources, nil
}

func (p *Provider) GetType() string {
	return p.typeName
}

func (p *Provider) GrantAccess(pc *domain.ProviderConfig, a *domain.Appeal) error {
	return nil
}

func (p *Provider) RevokeAccess(pc *domain.ProviderConfig, a *domain.Appeal) error {
	return nil
}

func (p *Provider) GetRoles(pc *domain.ProviderConfig, resourceType string) ([]*domain.Role, error) {
	return provider.GetRoles(pc, resourceType)
}

func (p *Provider) GetAccountTypes() []string {
	return []string{"user", "serviceAccount", "group", "domain"}
}
