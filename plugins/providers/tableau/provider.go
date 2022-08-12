package tableau

import (
	"github.com/mitchellh/mapstructure"
	pv "github.com/odpf/guardian/core/provider"
	"github.com/odpf/guardian/domain"
)

type provider struct {
	pv.PermissionManager

	typeName string
	Clients  map[string]TableauClient
	crypto   domain.Crypto
}

func NewProvider(typeName string, crypto domain.Crypto) *provider {
	return &provider{
		typeName: typeName,
		Clients:  map[string]TableauClient{},
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

	var resourceTypes []string
	for _, rc := range pc.Resources {
		resourceTypes = append(resourceTypes, rc.Type)
	}

	resources := []*domain.Resource{}

	if containsString(resourceTypes, ResourceTypeWorkbook) {
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
	}

	if containsString(resourceTypes, ResourceTypeFlow) {
		flows, err := client.GetFlows()
		if err != nil {
			return nil, err
		}
		for _, f := range flows {
			fl := f.ToDomain()
			fl.ProviderType = pc.Type
			fl.ProviderURN = pc.URN
			resources = append(resources, fl)
		}
	}

	if containsString(resourceTypes, ResourceTypeDataSource) {
		datasources, err := client.GetDataSources()
		if err != nil {
			return nil, err
		}
		for _, d := range datasources {
			ds := d.ToDomain()
			ds.ProviderType = pc.Type
			ds.ProviderURN = pc.URN
			resources = append(resources, ds)
		}
	}

	if containsString(resourceTypes, ResourceTypeView) {
		views, err := client.GetViews()
		if err != nil {
			return nil, err
		}
		for _, v := range views {
			vs := v.ToDomain()
			vs.ProviderType = pc.Type
			vs.ProviderURN = pc.URN
			resources = append(resources, vs)
		}
	}

	if containsString(resourceTypes, ResourceTypeMetric) {
		metrics, err := client.GetMetrics()
		if err != nil {
			return nil, err
		}
		for _, m := range metrics {
			mt := m.ToDomain()
			mt.ProviderType = pc.Type
			mt.ProviderURN = pc.URN
			resources = append(resources, mt)
		}
	}

	return resources, nil
}

func (p *provider) GrantAccess(pc *domain.ProviderConfig, a domain.Access) error {
	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return err
	}

	client, err := p.getClient(pc.URN, creds)
	if err != nil {
		return err
	}

	permissions := getPermissions(a)
	if a.Resource.Type == ResourceTypeWorkbook {
		w := new(Workbook)
		if err := w.FromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if p.Type == "" {
				if err := client.GrantWorkbookAccess(w, a.AccountID, p.Name); err != nil {
					return err
				}
			} else {
				if err := client.UpdateSiteRole(a.AccountID, p.Name); err != nil {
					return err
				}
			}
		}

		return nil
	} else if a.Resource.Type == ResourceTypeFlow {
		f := new(Flow)
		if err := f.FromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if p.Type == "" {
				if err := client.GrantFlowAccess(f, a.AccountID, p.Name); err != nil {
					return err
				}
			} else {
				if err := client.UpdateSiteRole(a.AccountID, p.Name); err != nil {
					return err
				}
			}
		}

		return nil
	} else if a.Resource.Type == ResourceTypeDataSource {
		d := new(DataSource)
		if err := d.FromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if p.Type == "" {
				if err := client.GrantDataSourceAccess(d, a.AccountID, p.Name); err != nil {
					return err
				}
			} else {
				if err := client.UpdateSiteRole(a.AccountID, p.Name); err != nil {
					return err
				}
			}
		}

		return nil
	} else if a.Resource.Type == ResourceTypeView {
		v := new(View)
		if err := v.FromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if p.Type == "" {
				if err := client.GrantViewAccess(v, a.AccountID, p.Name); err != nil {
					return err
				}
			} else {
				if err := client.UpdateSiteRole(a.AccountID, p.Name); err != nil {
					return err
				}
			}
		}

		return nil
	} else if a.Resource.Type == ResourceTypeMetric {
		m := new(Metric)
		if err := m.FromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if p.Type == "" {
				if err := client.GrantMetricAccess(m, a.AccountID, p.Name); err != nil {
					return err
				}
			} else {
				if err := client.UpdateSiteRole(a.AccountID, p.Name); err != nil {
					return err
				}
			}
		}

		return nil
	}

	return ErrInvalidResourceType
}

func (p *provider) RevokeAccess(pc *domain.ProviderConfig, a domain.Access) error {
	var creds Credentials
	if err := mapstructure.Decode(pc.Credentials, &creds); err != nil {
		return err
	}

	client, err := p.getClient(pc.URN, creds)
	if err != nil {
		return err
	}

	permissions := getPermissions(a)
	if a.Resource.Type == ResourceTypeWorkbook {
		w := new(Workbook)
		if err := w.FromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if p.Type == "" {
				if err := client.RevokeWorkbookAccess(w, a.AccountID, p.Name); err != nil {
					return err
				}
			}
		}

		if err := client.UpdateSiteRole(a.AccountID, "Unlicensed"); err != nil {
			return err
		}
		return nil
	} else if a.Resource.Type == ResourceTypeFlow {
		f := new(Flow)
		if err := f.FromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if p.Type == "" {
				if err := client.RevokeFlowAccess(f, a.AccountID, p.Name); err != nil {
					return err
				}
			}
		}

		if err := client.UpdateSiteRole(a.AccountID, "Unlicensed"); err != nil {
			return err
		}
		return nil
	} else if a.Resource.Type == ResourceTypeDataSource {
		d := new(DataSource)
		if err := d.FromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if p.Type == "" {
				if err := client.RevokeDataSourceAccess(d, a.AccountID, p.Name); err != nil {
					return err
				}
			}
		}

		if err := client.UpdateSiteRole(a.AccountID, "Unlicensed"); err != nil {
			return err
		}
		return nil
	} else if a.Resource.Type == ResourceTypeView {
		v := new(View)
		if err := v.FromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if p.Type == "" {
				if err := client.RevokeViewAccess(v, a.AccountID, p.Name); err != nil {
					return err
				}
			}
		}

		if err := client.UpdateSiteRole(a.AccountID, "Unlicensed"); err != nil {
			return err
		}
		return nil
	} else if a.Resource.Type == ResourceTypeMetric {
		m := new(Metric)
		if err := m.FromDomain(a.Resource); err != nil {
			return err
		}

		for _, p := range permissions {
			if p.Type == "" {
				if err := client.RevokeMetricAccess(m, a.AccountID, p.Name); err != nil {
					return err
				}
			}
		}

		if err := client.UpdateSiteRole(a.AccountID, "Unlicensed"); err != nil {
			return err
		}
		return nil
	}

	return ErrInvalidResourceType
}

func (p *provider) GetRoles(pc *domain.ProviderConfig, resourceType string) ([]*domain.Role, error) {
	return pv.GetRoles(pc, resourceType)
}

func (p *provider) GetAccountTypes() []string {
	return []string{
		AccountTypeUser,
	}
}

func (p *provider) getClient(providerURN string, credentials Credentials) (TableauClient, error) {
	if p.Clients[providerURN] != nil {
		return p.Clients[providerURN], nil
	}

	err := credentials.Decrypt(p.crypto)
	if err != nil {
		return nil, err
	}

	client, err := NewClient(&ClientConfig{
		Host:       credentials.Host,
		Username:   credentials.Username,
		Password:   credentials.Password,
		ContentURL: credentials.ContentURL,
	})
	if err != nil {
		return nil, err
	}

	p.Clients[providerURN] = client
	return client, nil
}

func getPermissions(a domain.Access) []Permission {
	var permissions []Permission
	for _, p := range a.Permissions {
		permissions = append(permissions, toPermission(p))
	}
	return permissions
}

func containsString(arr []string, v string) bool {
	for _, item := range arr {
		if item == v {
			return true
		}
	}
	return false
}
