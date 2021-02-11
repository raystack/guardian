package repositories

import "gorm.io/gorm"

// Repositories contains all repositories
type Repositories struct {
	Provider Provider
}

// New returns repositores
func New(db *gorm.DB) *Repositories {
	return &Repositories{
		Provider: NewProvider(db),
	}
}
