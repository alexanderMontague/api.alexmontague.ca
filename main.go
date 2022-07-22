package main

import (
	"api.alexmontague.ca/controllers"
	"api.alexmontague.ca/data"
	"api.alexmontague.ca/middleware"
	"fmt"
	"github.com/friendsofgo/graphiql"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"log"
	"net/http"
	"github.com/joho/godotenv"
)

const (
	// PORT - port server is using
	PORT = ":8088"
)

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)

	router.Use(middleware.LoggingMiddleware)

	// GraphiQL
	graphiqlHandler, err := graphiql.NewGraphiqlHandler("/graphql")
	if err != nil {
		fmt.Println("Error setting up GraphiQL %s", err)
	}

	// Routes
	router.HandleFunc("/", controllers.BaseURL).Methods("GET")
	router.HandleFunc("/email", controllers.EmailService).Methods("POST")
	router.HandleFunc("/resume", controllers.ResumeJSON).Methods("GET")
	router.HandleFunc("/graphql", controllers.GraphQL).Methods("POST")
	router.Handle("/graphiql", graphiqlHandler).Methods("GET")
	router.HandleFunc("/cors", controllers.CorsAnywhere).Methods("GET")

	// CORS middleware
	handler := cors.Default().Handler(router)

	// Start Server
	log.Fatal(http.ListenAndServe(PORT, handler))
}
func main() {
	fmt.Println("Running server on port", PORT)

	err := godotenv.Load(".env")

  if err != nil {
    fmt.Println("Error loading .env file")
  }

	data.SeedData()
	handleRequests()
}
