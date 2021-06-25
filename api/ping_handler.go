package api

import (
	"net/http"

	"github.com/odpf/guardian/utils"
)

// Ping handler
func Ping() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.ReturnJSON(w, "pong")
	}
}
