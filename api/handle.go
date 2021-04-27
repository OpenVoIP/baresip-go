package api

import (
	"encoding/json"
	"net/http"
)

func GetConfig(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "data": ""})
}
