package mocks

import (
	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewStore returns store mock
func NewStore() (*gorm.DB, sqlmock.Sqlmock, error) {
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqldb,
	}), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	return db, mock, nil
}
