package models

import (
	"time"

	"gorm.io/gorm"
)

// Provider database model
type Provider struct {
	ID        uint `gorm:"primaryKey"`
	Config    string
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
