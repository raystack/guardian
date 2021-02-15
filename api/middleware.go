package api

import (
	"net/http"
	"os"

	"github.com/gorilla/handlers"
)

func logger(next http.Handler) http.Handler {
	return handlers.LoggingHandler(os.Stdout, next)
}
