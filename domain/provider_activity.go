package domain

import "time"

type ProviderActivity struct {
	ID             string                 `json:"id" yaml:"id"`
	ProviderID     string                 `json:"provider_id" yaml:"provider_id"`
	ResourceID     string                 `json:"resource_id" yaml:"resource_id"`
	AccountType    string                 `json:"account_type" yaml:"account_type"`
	AccountID      string                 `json:"account_id" yaml:"account_id"`
	Timestamp      time.Time              `json:"timestamp" yaml:"timestamp"`
	Authorizations []string               `json:"authorizations" yaml:"authorizations"`
	Type           string                 `json:"type" yaml:"type"`
	Metadata       map[string]interface{} `json:"metadata" yaml:"metadata"`
	CreatedAt      time.Time              `json:"created_at" yaml:"created_at"`

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
