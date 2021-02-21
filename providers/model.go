package providers

import (
	"time"

	"github.com/odpf/guardian/domain"
	"gorm.io/gorm"
)

// Model is the database model for provider
type Model struct {
	ID        uint `gorm:"primaryKey"`
	Config    string
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName overrides the table name
func (Model) TableName() string {
	return "providers"
}

func (m *Model) fromDomain(p *domain.Provider) {
	m.ID = p.ID
	m.Config = p.Config
	m.CreatedAt = p.CreatedAt
	m.UpdatedAt = p.UpdatedAt
}

func (m *Model) toDomain() *domain.Provider {
	return &domain.Provider{
		ID:        m.ID,
		Config:    m.Config,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
