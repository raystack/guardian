package store

import "github.com/odpf/guardian/domain"

type AppealRepository interface {
	BulkUpsert([]*domain.Appeal) error
	Find(map[string]interface{}) ([]*domain.Appeal, error)
	GetByID(uint) (*domain.Appeal, error)
	Update(*domain.Appeal) error
}

type ApprovalRepository interface {
	BulkInsert([]*domain.Approval) error
	ListApprovals(*domain.ListApprovalsFilter) ([]*domain.Approval, error)
}

type PolicyRepository interface {
	Create(*domain.Policy) error
	Find() ([]*domain.Policy, error)
	GetOne(id string, version uint) (*domain.Policy, error)
}

type ProviderRepository interface {
	Create(*domain.Provider) error
	Update(*domain.Provider) error
	Find() ([]*domain.Provider, error)
	GetByID(uint) (*domain.Provider, error)
	GetOne(pType, urn string) (*domain.Provider, error)
	Delete(uint) error
}

type ResourceRepository interface {
	Find(filters map[string]interface{}) ([]*domain.Resource, error)
	GetOne(uint) (*domain.Resource, error)
	BulkUpsert([]*domain.Resource) error
	Update(*domain.Resource) error
}
