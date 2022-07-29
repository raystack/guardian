package postgres

import (
	"embed"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/odpf/guardian/internal/store"
	pg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var fs embed.FS

type Store struct {
	db *gorm.DB

	config *store.Config
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

	return &Store{gormDB, c}, nil
}

func (s *Store) DB() *gorm.DB {
	return s.db
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

	if err := m.Up(); errors.Is(err, migrate.ErrNoChange) {
		log.Println("migration schema version is up to date")
		return nil
	} else if err != nil {
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
