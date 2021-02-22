package utils

import (
	"encoding/json"
	"net/http"
)

// ReturnJSON writes JSON to the response body
func ReturnJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
