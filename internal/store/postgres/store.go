package postgres

import (
	"fmt"
	"log"

	"github.com/odpf/guardian/internal/store"
	"github.com/odpf/guardian/internal/store/postgres/model"
	pg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Store struct {
	db *gorm.DB
}

func NewStore(c *store.Config) (*Store, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s dbname=%s port=%s sslmode=%s password=%s",
		c.Host,
		c.User,
		c.Name,
		c.Port,
		c.SslMode,
		c.Password,
	)

	gormDB, err := gorm.Open(pg.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panic(err)
	}

	return &Store{gormDB}, nil
}

func (s *Store) DB() *gorm.DB {
	return s.db
}

func (s *Store) Migrate() error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error; err != nil {
			return err
		}

		return tx.AutoMigrate(
			&model.Provider{},
			&model.Policy{},
			&model.Resource{},
			&model.Appeal{},
			&model.Approval{},
			&model.Approver{},
		)
	})
}
