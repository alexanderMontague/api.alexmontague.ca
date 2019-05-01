package controllers

import (
	"am.ca_server/helpers"
	"encoding/json"
	"fmt"
	"net/http"
)

// BaseURL Route
// Route : '/'
// Type  : 'GET'
func BaseURL(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 404, Message: "Invalid Route"})
}

// EmailService Route
// Route : '/email
// Type  : 'POST'
func EmailService(w http.ResponseWriter, r *http.Request) {
	fmt.Println("request", r.Body)

	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(helpers.Response{Error: false, Code: 200, Message: "Email Received"})
}
