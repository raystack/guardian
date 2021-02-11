package handlers

import (
	"net/http"
)

// Ping handler
func Ping() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		returnJSON(w, "pong")
		return
	}
}
