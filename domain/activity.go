package domain

import (
	"errors"
	"time"
)

type Activity struct {
	ID                 string                 `json:"id" yaml:"id"`
	ProviderID         string                 `json:"provider_id" yaml:"provider_id"`
	ResourceID         string                 `json:"resource_id" yaml:"resource_id"`
	ProviderActivityID string                 `json:"provider_activity_id" yaml:"provider_activity_id"`
	AccountType        string                 `json:"account_type" yaml:"account_type"`
	AccountID          string                 `json:"account_id" yaml:"account_id"`
	Timestamp          time.Time              `json:"timestamp" yaml:"timestamp"`
	Authorizations     []string               `json:"authorizations" yaml:"authorizations"`
	RelatedPermissions []string               `json:"related_permissions" yaml:"related_permissions"`
	Type               string                 `json:"type" yaml:"type"`
	Metadata           map[string]interface{} `json:"metadata" yaml:"metadata"`
	CreatedAt          time.Time              `json:"created_at" yaml:"created_at"`

	Provider *Provider `json:"provider,omitempty" yaml:"provider,omitempty"`
	Resource *Resource `json:"resource,omitempty" yaml:"resource,omitempty"`
}

type ListProviderActivitiesFilter struct {
	ProviderIDs  []string
	ResourceIDs  []string
	AccountIDs   []string
	Types        []string
	TimestampGte *time.Time
	TimestampLte *time.Time
}

type ImportActivitiesFilter struct {
	ProviderID   string
	ResourceIDs  []string
	AccountIDs   []string
	TimestampGte *time.Time
	TimestampLte *time.Time

	resources map[string]*Resource
}

func (f *ImportActivitiesFilter) PopulateResources(resources map[string]*Resource) error {
	if f.ResourceIDs == nil {
		return nil
	}
	if resources == nil {
		return errors.New("resources cannot be nil")
	}

	f.resources = make(map[string]*Resource, len(f.ResourceIDs))
	for _, resourceID := range f.ResourceIDs {
		resource, ok := resources[resourceID]
		if !ok {
			return errors.New("resource not found")
		}
		f.resources[resourceID] = resource
	}

	return nil
}

func (f *ImportActivitiesFilter) GetResources() []*Resource {
	if f.resources == nil {
		return nil
	}

	resources := make([]*Resource, 0, len(f.resources))
	for _, resource := range f.resources {
		resources = append(resources, resource)
	}

	return resources
}
