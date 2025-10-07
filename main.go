package main

import (
	"fmt"
	"log"
	"net/http"

	"api.alexmontague.ca/controllers"
	"api.alexmontague.ca/data"
	"api.alexmontague.ca/internal/cron"
	"api.alexmontague.ca/internal/database"
	"api.alexmontague.ca/internal/nhl/service"
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

	router.HandleFunc("/budget/categories", controllers.GetCategories).Methods("GET")
	router.HandleFunc("/budget/categories", controllers.CreateCategory).Methods("POST")
	router.HandleFunc("/budget/categories/{id}", controllers.UpdateCategory).Methods("PUT")
	router.HandleFunc("/budget/categories/{id}", controllers.DeleteCategory).Methods("DELETE")
	router.HandleFunc("/budget/categories", controllers.DeleteAllCategories).Methods("DELETE")

	router.HandleFunc("/budget/budgets", controllers.GetBudgets).Methods("GET")
	router.HandleFunc("/budget/budgets", controllers.CreateBudget).Methods("POST")
	router.HandleFunc("/budget/budgets/{id}", controllers.UpdateBudget).Methods("PUT")
	router.HandleFunc("/budget/budgets/{id}", controllers.DeleteBudget).Methods("DELETE")
	router.HandleFunc("/budget/budgets", controllers.DeleteAllBudgets).Methods("DELETE")

	router.HandleFunc("/budget/transactions", controllers.GetTransactions).Methods("GET")
	router.HandleFunc("/budget/transactions", controllers.CreateTransactions).Methods("POST")
	router.HandleFunc("/budget/transactions/{id}", controllers.UpdateTransaction).Methods("PATCH")
	router.HandleFunc("/budget/transactions/{id}", controllers.DeleteTransaction).Methods("DELETE")
	router.HandleFunc("/budget/transactions", controllers.DeleteAllTransactions).Methods("DELETE")

	router.HandleFunc("/budget/clear", controllers.ClearAllBudgetData).Methods("DELETE")
	router.HandleFunc("/budget/export", controllers.ExportBudgetData).Methods("GET")
	router.HandleFunc("/budget/import", controllers.ImportBudgetData).Methods("POST")

	// CORS middleware
	handler := cors.Default().Handler(router)

	// Start Server
	log.Fatal(http.ListenAndServe(PORT, handler))
}

func main() {
	// Initialize database
	if err := database.InitDB(database.DB_PATH); err != nil {
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

	// Initialize model prediction system
	go func() {
		// Ensure models are initialized on startup
		service.InitializeModels()
		// Set the active model version to use for predictions
		service.SetActiveModelVersion(1) // Using the original model as default
	}()

	handleRequests()
}
