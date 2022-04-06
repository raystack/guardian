package audit

import (
	"context"

	"gorm.io/gorm"
)

type auditModel struct {
}

type postgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository() *postgresRepository {
	return &postgresRepository{}
}

func (r *postgresRepository) Insert(ctx context.Context, l *Log) error {
	// TODO: insert to db
	return nil
}
