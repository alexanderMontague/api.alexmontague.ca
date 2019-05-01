package main

import (
	"am.ca_server/controllers"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

const (
	// PORT : port that server is hosted under
	PORT = ":8080"
)

// Response : format of all responses coming back from server
type Response struct {
	Error   bool   `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", controllers.BaseURL).Methods("GET")

	log.Fatal(http.ListenAndServe(PORT, router))
}

func main() {
	fmt.Println("Running server on port", PORT)

	handleRequests()
}
