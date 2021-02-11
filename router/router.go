package router

import (
	"github.com/gorilla/mux"
	"github.com/odpf/guardian/handlers"
	"github.com/odpf/guardian/usecases"
)

// New instantiates the service router
func New(u *usecases.UseCases) *mux.Router {
	r := mux.NewRouter().StrictSlash(true)

	r.Use(logger)
	registerRoutes(r)

	return r
}

func registerRoutes(r *mux.Router) {
	r.Methods("GET").Path("/ping").Handler(handlers.Ping())
}
