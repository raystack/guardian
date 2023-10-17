package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/raystack/guardian/pkg/auth"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/raystack/guardian/internal/store"
	pg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var fs embed.FS

type Store struct {
	db *gorm.DB

	config *store.Config
}

const (
	namespaceRLSSetQuery   = "SET app.current_tenant = '%s'"
	namespaceRLSResetQuery = "RESET app.current_tenant"
)

type Connection interface {
	Tx(ctx context.Context, fc func(tx *gorm.DB) error) error
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
	if strings.ToLower(c.LogLevel) == "debug" {
		gormDB = gormDB.Debug()
	}

	// Get generic database object sql.DB to use its functions
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Panic(err)
	}

	// set connection pool settings only if user provided them
	// otherwise default to gorm settings
	if c.MaxIdleConns != 0 {
		sqlDB.SetMaxIdleConns(c.MaxIdleConns)
	}

	if c.MaxOpenConns != 0 {
		sqlDB.SetMaxOpenConns(c.MaxOpenConns)
	}

	if c.ConnMaxLifetimeInMs != 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(c.ConnMaxLifetimeInMs) * time.Millisecond)
	}

	if c.ConnMaxIdleTimeInMs != 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(c.ConnMaxIdleTimeInMs) * time.Millisecond)
	}

	return &Store{gormDB, c}, nil
}

func (s *Store) DB() *gorm.DB {
	return s.db
}

func (s *Store) Tx(ctx context.Context, fc func(tx *gorm.DB) error) error {
	return s.db.WithContext(ctx).Connection(func(conn *gorm.DB) error {
		return conn.Transaction(func(tx *gorm.DB) error {
			// set tenant context
			if err := tx.Exec(fmt.Sprintf(namespaceRLSSetQuery, namespaceFromContext(ctx))).Error; err != nil {
				return err
			}

			// execute the requested operation
			fcErr := fc(tx)

			// reset tenant context
			_ = tx.Exec(namespaceRLSResetQuery).Error
			return fcErr
		})
	})
}

func (s *Store) Migrate() error {
	iofsDriver, err := iofs.New(fs, "migrations")
	if err != nil {
		return err
	}
	m, err := migrate.NewWithSourceInstance("iofs", iofsDriver, toConnectionString(s.config))
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("migration schema version is up to date")
			return nil
		}
		return err
	}

	return nil
}

func toConnectionString(c *store.Config) string {
	pgURL := &url.URL{
		Scheme: "postgres",
		Host:   net.JoinHostPort(c.Host, c.Port),
		User:   url.UserPassword(c.User, c.Password),
		Path:   c.Name,
	}
	q := pgURL.Query()
	if c.SslMode != "" {
		q.Add("sslmode", c.SslMode)
	}
	pgURL.RawQuery = q.Encode()

	return pgURL.String()
}

func namespaceFromContext(ctx context.Context) uuid.UUID {
	return auth.FetchNamespace(ctx)
}
