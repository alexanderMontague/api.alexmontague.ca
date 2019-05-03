package main

import (
	"am.ca-server/controllers"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"log"
	"net/http"
)

const (
	// PORT : port that server is hosted under
	PORT = ":8088"
)

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", controllers.BaseURL).Methods("GET")
	router.HandleFunc("/email", controllers.EmailService).Methods("POST")

	log.Fatal(http.ListenAndServe(PORT, router))
}

func main() {
	fmt.Println("Running server on port", PORT)

	// load dotenv variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	handleRequests()
}
