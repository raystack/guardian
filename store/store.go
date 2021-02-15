package store

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Config for database connection
type Config struct {
	Host     string
	User     string
	Password string
	Name     string
	Port     string
	SslMode  string
}

// New returns the database instance
func New(c *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s dbname=%s password=%s port=%s sslmode=%s",
		c.Host,
		c.User,
		c.Name,
		c.Password,
		c.Port,
		c.SslMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Panic(err)
	}

	return db, err
}

// Migrate auto migrate models
func Migrate(db *gorm.DB, models ...interface{}) error {
	return db.AutoMigrate(models...)
}
