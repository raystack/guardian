package api

import (
	"github.com/gorilla/mux"
	"github.com/odpf/guardian/api/handlers"
	"github.com/purini-to/zapmw"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New initializes the service router
func New(logger *zap.Logger) *mux.Router {
	r := mux.NewRouter().StrictSlash(true)

	zapMiddlewares := []mux.MiddlewareFunc{
		zapmw.WithZap(logger),
		zapmw.Request(zapcore.InfoLevel, "request"),
		zapmw.Recoverer(zapcore.ErrorLevel, "recover", zapmw.RecovererDefault),
	}
	r.Use(zapMiddlewares...)

	r.Methods("GET").Path("/ping").Handler(handlers.Ping())

	return r
}
