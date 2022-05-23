package store

import "github.com/odpf/guardian/domain"

type AppealRepository interface {
	BulkUpsert([]*domain.Appeal) error
	Find(*domain.ListAppealsFilter) ([]*domain.Appeal, error)
	GetByID(id string) (*domain.Appeal, error)
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
	GetByID(id string) (*domain.Provider, error)
	GetTypes() ([]domain.ProviderType, error)
	GetOne(pType, urn string) (*domain.Provider, error)
	Delete(id string) error
}

type ResourceRepository interface {
	Find(filters map[string]interface{}) ([]*domain.Resource, error)
	GetOne(id string) (*domain.Resource, error)
	BulkUpsert([]*domain.Resource) error
	Update(*domain.Resource) error
	Delete(id string) error
}
