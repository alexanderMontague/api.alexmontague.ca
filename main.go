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

	router.HandleFunc("/auth/register", controllers.Register).Methods("POST")
	router.HandleFunc("/auth/login", controllers.Login).Methods("POST")

	budgetRouter := router.PathPrefix("/budget").Subrouter()
	budgetRouter.Use(middleware.AuthMiddleware)

	budgetRouter.HandleFunc("/me", controllers.GetCurrentUser).Methods("GET")

	budgetRouter.HandleFunc("/categories", controllers.GetCategories).Methods("GET")
	budgetRouter.HandleFunc("/categories", controllers.CreateCategory).Methods("POST")
	budgetRouter.HandleFunc("/categories/{id}", controllers.UpdateCategory).Methods("PUT")
	budgetRouter.HandleFunc("/categories/{id}", controllers.DeleteCategory).Methods("DELETE")
	budgetRouter.HandleFunc("/categories", controllers.DeleteAllCategories).Methods("DELETE")

	budgetRouter.HandleFunc("/budgets", controllers.GetBudgets).Methods("GET")
	budgetRouter.HandleFunc("/budgets", controllers.CreateBudget).Methods("POST")
	budgetRouter.HandleFunc("/budgets/{id}", controllers.UpdateBudget).Methods("PUT")
	budgetRouter.HandleFunc("/budgets/{id}", controllers.DeleteBudget).Methods("DELETE")
	budgetRouter.HandleFunc("/budgets", controllers.DeleteAllBudgets).Methods("DELETE")

	budgetRouter.HandleFunc("/transactions", controllers.GetTransactions).Methods("GET")
	budgetRouter.HandleFunc("/transactions", controllers.CreateTransactions).Methods("POST")
	budgetRouter.HandleFunc("/transactions/{id}", controllers.UpdateTransaction).Methods("PATCH")
	budgetRouter.HandleFunc("/transactions/{id}", controllers.DeleteTransaction).Methods("DELETE")
	budgetRouter.HandleFunc("/transactions", controllers.DeleteAllTransactions).Methods("DELETE")

	budgetRouter.HandleFunc("/clear", controllers.ClearAllBudgetData).Methods("DELETE")
	budgetRouter.HandleFunc("/export", controllers.ExportBudgetData).Methods("GET")
	budgetRouter.HandleFunc("/import", controllers.ImportBudgetData).Methods("POST")

	// CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	handler := c.Handler(router)

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
