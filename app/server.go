package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/odpf/guardian/config"
	"github.com/odpf/guardian/database"
	"github.com/odpf/guardian/models"
	"github.com/odpf/guardian/repositories"
	"github.com/odpf/guardian/router"
	"github.com/odpf/guardian/usecases"
)

// RunServer runs the application server
func RunServer(c *config.Config) error {
	db, err := database.New(&database.Config{
		Host:     c.DBHost,
		User:     c.DBUser,
		Password: c.DBPassword,
		Name:     c.DBName,
		Port:     c.DBPort,
		SslMode:  c.DBSslMode,
	})
	if err != nil {
		return err
	}

	models := []interface{}{
		&models.Provider{},
	}
	database.Migrate(db, models...)

	allRepositories := repositories.New(db)
	allUsecases := usecases.New(allRepositories)

	r := router.New(allUsecases)

	log.Printf("running server on port %d\n", c.Port)
	return http.ListenAndServe(fmt.Sprintf(":%d", c.Port), r)
}
