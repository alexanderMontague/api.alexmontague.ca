package main

import (
	"fmt"
	"log"
	"net/http"

	"api.alexmontague.ca/controllers"
	"api.alexmontague.ca/data"
	"api.alexmontague.ca/internal/cron"
	"api.alexmontague.ca/internal/database"
	"api.alexmontague.ca/middleware"
	"github.com/friendsofgo/graphiql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
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
		fmt.Printf("Error setting up GraphiQL %s", err)
	}

	// Routes
	router.HandleFunc("/", controllers.BaseURL).Methods("GET")
	router.HandleFunc("/email", controllers.EmailService).Methods("POST")
	router.HandleFunc("/resume", controllers.ResumeJSON).Methods("GET")
	router.HandleFunc("/graphql", controllers.GraphQL).Methods("POST")
	router.Handle("/graphiql", graphiqlHandler).Methods("GET")
	router.HandleFunc("/cors", controllers.CorsAnywhere).Methods("GET")
	router.HandleFunc("/nhl/shots", controllers.GetPlayerShotStats).Methods("GET")
	router.HandleFunc("/nhl/shots/records", controllers.GetPlayerShotRecords).Methods("GET")
	router.HandleFunc("/nhl/shots/seed", controllers.SeedAndValidatePredictions).Methods("GET")

	// CORS middleware
	handler := cors.Default().Handler(router)

	// Start Server
	log.Fatal(http.ListenAndServe(PORT, handler))
}

func main() {
	// Initialize database
	if err := database.InitDB("./nhl_predictions.db"); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Start cron scheduler
	cron.StartScheduler()
	defer cron.StopScheduler()

	fmt.Println("Running server on port", PORT)

	err := godotenv.Load(".env")

	if err != nil {
		fmt.Println("Error loading .env file")
	}

	data.SeedData()
	handleRequests()
}
