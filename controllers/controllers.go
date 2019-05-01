package controllers

import (
	"am.ca_server/helpers"
	"encoding/json"
	"fmt"
	"net/http"
)

func BaseURL(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(datatypes.Response{Error: true, Code: 404, Message: "Invalid Route"})
}
