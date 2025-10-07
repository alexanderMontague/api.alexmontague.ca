package controllers

import (
	"encoding/json"
	"net/http"

	"api.alexmontague.ca/helpers"
	"api.alexmontague.ca/internal/database/repository"
	"github.com/gorilla/mux"
)

func GetCategories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	categories, err := repository.GetCategories()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 500, Message: "Failed to fetch categories"})
		return
	}

	json.NewEncoder(w).Encode(categories)
}

func CreateCategory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	var category repository.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 400, Message: "Invalid request body"})
		return
	}

	saved, err := repository.SaveCategory(category)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 500, Message: "Failed to save category"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(saved)
}

func UpdateCategory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	vars := mux.Vars(r)
	id := vars["id"]

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 400, Message: "Invalid request body"})
		return
	}

	updated, err := repository.UpdateCategory(id, updates)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 500, Message: "Failed to update category"})
		return
	}

	json.NewEncoder(w).Encode(updated)
}

func DeleteCategory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	vars := mux.Vars(r)
	id := vars["id"]

	if err := repository.DeleteCategory(id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 500, Message: "Failed to delete category"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteAllCategories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	if err := repository.DeleteAllCategories(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 500, Message: "Failed to delete all categories"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func GetBudgets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	budgets, err := repository.GetBudgets()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 500, Message: "Failed to fetch budgets"})
		return
	}

	json.NewEncoder(w).Encode(budgets)
}

func CreateBudget(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	var budget repository.Budget
	if err := json.NewDecoder(r.Body).Decode(&budget); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 400, Message: "Invalid request body"})
		return
	}

	saved, err := repository.SaveBudget(budget)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 500, Message: "Failed to save budget"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(saved)
}

func UpdateBudget(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	vars := mux.Vars(r)
	id := vars["id"]

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 400, Message: "Invalid request body"})
		return
	}

	updated, err := repository.UpdateBudget(id, updates)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 500, Message: "Failed to update budget"})
		return
	}

	json.NewEncoder(w).Encode(updated)
}

func DeleteBudget(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	vars := mux.Vars(r)
	id := vars["id"]

	if err := repository.DeleteBudget(id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 500, Message: "Failed to delete budget"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteAllBudgets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	if err := repository.DeleteAllBudgets(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 500, Message: "Failed to delete all budgets"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func GetTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	transactions, err := repository.GetTransactions()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 500, Message: "Failed to fetch transactions"})
		return
	}

	json.NewEncoder(w).Encode(transactions)
}

func CreateTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	var transactions []repository.Transaction
	if err := json.NewDecoder(r.Body).Decode(&transactions); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 400, Message: "Invalid request body"})
		return
	}

	saved, err := repository.SaveTransactions(transactions)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 500, Message: "Failed to save transactions"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(saved)
}

func UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	vars := mux.Vars(r)
	id := vars["id"]

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 400, Message: "Invalid request body"})
		return
	}

	updated, err := repository.UpdateTransaction(id, updates)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 500, Message: "Failed to update transaction"})
		return
	}

	json.NewEncoder(w).Encode(updated)
}

func DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	vars := mux.Vars(r)
	id := vars["id"]

	if err := repository.DeleteTransaction(id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 500, Message: "Failed to delete transaction"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteAllTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	if err := repository.DeleteAllTransactions(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 500, Message: "Failed to delete all transactions"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ClearAllBudgetData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	if err := repository.ClearAllBudgetData(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 500, Message: "Failed to clear all data"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ExportBudgetData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	data, err := repository.ExportBudgetData()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 500, Message: "Failed to export data"})
		return
	}

	w.Write([]byte(data))
}

func ImportBudgetData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 400, Message: "Invalid request body"})
		return
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 400, Message: "Failed to process data"})
		return
	}

	if err := repository.ImportBudgetData(string(jsonData)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 500, Message: "Failed to import data"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
