package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"api.alexmontague.ca/helpers"
	"api.alexmontague.ca/internal/database/repository"
	"api.alexmontague.ca/middleware"
	"github.com/gorilla/mux"
)

func logAndRespondError(w http.ResponseWriter, statusCode int, message string, err error) {
	log.Printf("[Budget Error] %s: %v", message, err)
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: statusCode, Message: message})
}

func GetCategories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	authUser := middleware.GetAuthUser(r)
	categories, err := repository.GetCategories(authUser.UserID)
	if err != nil {
		logAndRespondError(w, http.StatusInternalServerError, "Failed to fetch categories", err)
		return
	}

	json.NewEncoder(w).Encode(categories)
}

func CreateCategory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	authUser := middleware.GetAuthUser(r)
	var category repository.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		logAndRespondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	saved, err := repository.SaveCategory(category, authUser.UserID)
	if err != nil {
		logAndRespondError(w, http.StatusInternalServerError, "Failed to save category", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(saved)
}

func UpdateCategory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	authUser := middleware.GetAuthUser(r)
	vars := mux.Vars(r)
	id := vars["id"]

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		logAndRespondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	updated, err := repository.UpdateCategory(id, authUser.UserID, updates)
	if err != nil {
		logAndRespondError(w, http.StatusInternalServerError, "Failed to update category", err)
		return
	}

	json.NewEncoder(w).Encode(updated)
}

func DeleteCategory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	authUser := middleware.GetAuthUser(r)
	vars := mux.Vars(r)
	id := vars["id"]

	if err := repository.DeleteCategory(id, authUser.UserID); err != nil {
		logAndRespondError(w, http.StatusInternalServerError, "Failed to delete category", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteAllCategories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	authUser := middleware.GetAuthUser(r)
	if err := repository.DeleteAllCategories(authUser.UserID); err != nil {
		logAndRespondError(w, http.StatusInternalServerError, "Failed to delete all categories", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func GetBudgets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	authUser := middleware.GetAuthUser(r)
	budgets, err := repository.GetBudgets(authUser.UserID)
	if err != nil {
		logAndRespondError(w, http.StatusInternalServerError, "Failed to fetch budgets", err)
		return
	}

	json.NewEncoder(w).Encode(budgets)
}

func CreateBudget(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	authUser := middleware.GetAuthUser(r)
	var budget repository.Budget
	if err := json.NewDecoder(r.Body).Decode(&budget); err != nil {
		logAndRespondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	saved, err := repository.SaveBudget(budget, authUser.UserID)
	if err != nil {
		logAndRespondError(w, http.StatusInternalServerError, "Failed to save budget", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(saved)
}

func UpdateBudget(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	authUser := middleware.GetAuthUser(r)
	vars := mux.Vars(r)
	id := vars["id"]

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		logAndRespondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	updated, err := repository.UpdateBudget(id, authUser.UserID, updates)
	if err != nil {
		logAndRespondError(w, http.StatusInternalServerError, "Failed to update budget", err)
		return
	}

	json.NewEncoder(w).Encode(updated)
}

func DeleteBudget(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	authUser := middleware.GetAuthUser(r)
	vars := mux.Vars(r)
	id := vars["id"]

	if err := repository.DeleteBudget(id, authUser.UserID); err != nil {
		logAndRespondError(w, http.StatusInternalServerError, "Failed to delete budget", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteAllBudgets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	authUser := middleware.GetAuthUser(r)
	if err := repository.DeleteAllBudgets(authUser.UserID); err != nil {
		logAndRespondError(w, http.StatusInternalServerError, "Failed to delete all budgets", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func GetTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	authUser := middleware.GetAuthUser(r)
	transactions, err := repository.GetTransactions(authUser.UserID)
	if err != nil {
		logAndRespondError(w, http.StatusInternalServerError, "Failed to fetch transactions", err)
		return
	}

	json.NewEncoder(w).Encode(transactions)
}

func CreateTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	authUser := middleware.GetAuthUser(r)
	var transactions []repository.Transaction
	if err := json.NewDecoder(r.Body).Decode(&transactions); err != nil {
		logAndRespondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	saved, err := repository.SaveTransactions(transactions, authUser.UserID)
	if err != nil {
		logAndRespondError(w, http.StatusInternalServerError, "Failed to save transactions", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(saved)
}

func UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	authUser := middleware.GetAuthUser(r)
	vars := mux.Vars(r)
	id := vars["id"]

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		logAndRespondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	updated, err := repository.UpdateTransaction(id, authUser.UserID, updates)
	if err != nil {
		logAndRespondError(w, http.StatusInternalServerError, "Failed to update transaction", err)
		return
	}

	json.NewEncoder(w).Encode(updated)
}

func DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	authUser := middleware.GetAuthUser(r)
	vars := mux.Vars(r)
	id := vars["id"]

	if err := repository.DeleteTransaction(id, authUser.UserID); err != nil {
		logAndRespondError(w, http.StatusInternalServerError, "Failed to delete transaction", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteAllTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	authUser := middleware.GetAuthUser(r)
	if err := repository.DeleteAllTransactions(authUser.UserID); err != nil {
		logAndRespondError(w, http.StatusInternalServerError, "Failed to delete all transactions", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ClearAllBudgetData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	authUser := middleware.GetAuthUser(r)
	if err := repository.ClearAllBudgetData(authUser.UserID); err != nil {
		logAndRespondError(w, http.StatusInternalServerError, "Failed to clear all data", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ExportBudgetData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	authUser := middleware.GetAuthUser(r)
	data, err := repository.ExportBudgetData(authUser.UserID)
	if err != nil {
		logAndRespondError(w, http.StatusInternalServerError, "Failed to export data", err)
		return
	}

	w.Write([]byte(data))
}

func ImportBudgetData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	authUser := middleware.GetAuthUser(r)
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		logAndRespondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		logAndRespondError(w, http.StatusBadRequest, "Failed to process data", err)
		return
	}

	if err := repository.ImportBudgetData(string(jsonData), authUser.UserID); err != nil {
		logAndRespondError(w, http.StatusInternalServerError, "Failed to import data", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
