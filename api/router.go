package api

import (
	"github.com/gorilla/mux"
	"github.com/odpf/guardian/api/handlers"
)

// New initializes the service router
func New() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)

	r.Use(logger)

	r.Methods("GET").Path("/ping").Handler(handlers.Ping())

	return r
}
