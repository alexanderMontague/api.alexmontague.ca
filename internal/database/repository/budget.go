package repository

import (
	"database/sql"
	"encoding/json"
	"time"

	"api.alexmontague.ca/internal/database"
)

type Category struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	MonthlyBudget *float64 `json:"monthlyBudget,omitempty"`
	Color         *string  `json:"color,omitempty"`
	CreatedAt     string   `json:"createdAt"`
	UpdatedAt     string   `json:"updatedAt"`
}

type Budget struct {
	ID                string            `json:"id"`
	Month             string            `json:"month"`
	Allocations       map[string]float64 `json:"allocations"`
	AvailableToBudget float64           `json:"availableToBudget"`
	CreatedAt         string            `json:"createdAt"`
	UpdatedAt         string            `json:"updatedAt"`
}

type Transaction struct {
	ID              string  `json:"id"`
	BudgetID        string  `json:"budgetId"`
	CategoryID      *string `json:"categoryId,omitempty"`
	TransactionHash string  `json:"transactionHash"`
	Date            string  `json:"date"`
	Merchant        string  `json:"merchant"`
	Amount          float64 `json:"amount"`
	Description     string  `json:"description"`
	AccountType     string  `json:"accountType"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

func GetCategories() ([]Category, error) {
	rows, err := database.DB.Query(`
		SELECT id, name, monthly_budget, color, created_at, updated_at
		FROM categories
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []Category{}
	for rows.Next() {
		var c Category
		err := rows.Scan(&c.ID, &c.Name, &c.MonthlyBudget, &c.Color, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

func SaveCategory(category Category) (*Category, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	category.CreatedAt = now
	category.UpdatedAt = now

	_, err := database.DB.Exec(`
		INSERT INTO categories (id, name, monthly_budget, color, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, category.ID, category.Name, category.MonthlyBudget, category.Color, category.CreatedAt, category.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &category, nil
}

func UpdateCategory(id string, updates map[string]interface{}) (*Category, error) {
	now := time.Now().Format("2006-01-02 15:04:05")

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var category Category
	err = tx.QueryRow(`
		SELECT id, name, monthly_budget, color, created_at, updated_at
		FROM categories WHERE id = ?
	`, id).Scan(&category.ID, &category.Name, &category.MonthlyBudget, &category.Color, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if name, ok := updates["name"].(string); ok {
		category.Name = name
	}
	if monthlyBudget, ok := updates["monthlyBudget"].(float64); ok {
		category.MonthlyBudget = &monthlyBudget
	}
	if color, ok := updates["color"].(string); ok {
		category.Color = &color
	}
	category.UpdatedAt = now

	_, err = tx.Exec(`
		UPDATE categories
		SET name = ?, monthly_budget = ?, color = ?, updated_at = ?
		WHERE id = ?
	`, category.Name, category.MonthlyBudget, category.Color, category.UpdatedAt, id)

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &category, nil
}

func DeleteCategory(id string) error {
	_, err := database.DB.Exec("DELETE FROM categories WHERE id = ?", id)
	return err
}

func DeleteAllCategories() error {
	_, err := database.DB.Exec("DELETE FROM categories")
	return err
}

func GetBudgets() ([]Budget, error) {
	rows, err := database.DB.Query(`
		SELECT id, month, allocations, available_to_budget, created_at, updated_at
		FROM budgets
		ORDER BY month DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	budgets := []Budget{}
	for rows.Next() {
		var b Budget
		var allocationsJSON string
		err := rows.Scan(&b.ID, &b.Month, &allocationsJSON, &b.AvailableToBudget, &b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(allocationsJSON), &b.Allocations); err != nil {
			return nil, err
		}
		budgets = append(budgets, b)
	}
	return budgets, rows.Err()
}

func SaveBudget(budget Budget) (*Budget, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	budget.CreatedAt = now
	budget.UpdatedAt = now

	allocationsJSON, err := json.Marshal(budget.Allocations)
	if err != nil {
		return nil, err
	}

	_, err = database.DB.Exec(`
		INSERT INTO budgets (id, month, allocations, available_to_budget, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, budget.ID, budget.Month, string(allocationsJSON), budget.AvailableToBudget, budget.CreatedAt, budget.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &budget, nil
}

func UpdateBudget(id string, updates map[string]interface{}) (*Budget, error) {
	now := time.Now().Format("2006-01-02 15:04:05")

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var budget Budget
	var allocationsJSON string
	err = tx.QueryRow(`
		SELECT id, month, allocations, available_to_budget, created_at, updated_at
		FROM budgets WHERE id = ?
	`, id).Scan(&budget.ID, &budget.Month, &allocationsJSON, &budget.AvailableToBudget, &budget.CreatedAt, &budget.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(allocationsJSON), &budget.Allocations); err != nil {
		return nil, err
	}

	if month, ok := updates["month"].(string); ok {
		budget.Month = month
	}
	if allocations, ok := updates["allocations"].(map[string]interface{}); ok {
		alloc := make(map[string]float64)
		for k, v := range allocations {
			if val, ok := v.(float64); ok {
				alloc[k] = val
			}
		}
		budget.Allocations = alloc
	}
	if availableToBudget, ok := updates["availableToBudget"].(float64); ok {
		budget.AvailableToBudget = availableToBudget
	}
	budget.UpdatedAt = now

	newAllocationsJSON, err := json.Marshal(budget.Allocations)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`
		UPDATE budgets
		SET month = ?, allocations = ?, available_to_budget = ?, updated_at = ?
		WHERE id = ?
	`, budget.Month, string(newAllocationsJSON), budget.AvailableToBudget, budget.UpdatedAt, id)

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &budget, nil
}

func DeleteBudget(id string) error {
	_, err := database.DB.Exec("DELETE FROM budgets WHERE id = ?", id)
	return err
}

func DeleteAllBudgets() error {
	_, err := database.DB.Exec("DELETE FROM budgets")
	return err
}

func GetTransactions() ([]Transaction, error) {
	rows, err := database.DB.Query(`
		SELECT id, budget_id, category_id, transaction_hash, date, merchant, amount, description, account_type, created_at, updated_at
		FROM transactions
		ORDER BY date DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := []Transaction{}
	for rows.Next() {
		var t Transaction
		var categoryID sql.NullString
		err := rows.Scan(&t.ID, &t.BudgetID, &categoryID, &t.TransactionHash, &t.Date, &t.Merchant, &t.Amount, &t.Description, &t.AccountType, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if categoryID.Valid {
			t.CategoryID = &categoryID.String
		}
		transactions = append(transactions, t)
	}
	return transactions, rows.Err()
}

func SaveTransactions(transactions []Transaction) ([]Transaction, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO transactions (id, budget_id, category_id, transaction_hash, date, merchant, amount, description, account_type, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	now := time.Now().Format("2006-01-02 15:04:05")
	savedTransactions := []Transaction{}

	for _, t := range transactions {
		t.CreatedAt = now
		t.UpdatedAt = now

		_, err = stmt.Exec(t.ID, t.BudgetID, t.CategoryID, t.TransactionHash, t.Date, t.Merchant, t.Amount, t.Description, t.AccountType, t.CreatedAt, t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		savedTransactions = append(savedTransactions, t)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return savedTransactions, nil
}

func UpdateTransaction(id string, updates map[string]interface{}) (*Transaction, error) {
	now := time.Now().Format("2006-01-02 15:04:05")

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var t Transaction
	var categoryID sql.NullString
	err = tx.QueryRow(`
		SELECT id, budget_id, category_id, transaction_hash, date, merchant, amount, description, account_type, created_at, updated_at
		FROM transactions WHERE id = ?
	`, id).Scan(&t.ID, &t.BudgetID, &categoryID, &t.TransactionHash, &t.Date, &t.Merchant, &t.Amount, &t.Description, &t.AccountType, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if categoryID.Valid {
		t.CategoryID = &categoryID.String
	}

	if budgetID, ok := updates["budgetId"].(string); ok {
		t.BudgetID = budgetID
	}
	if categoryIDUpdate, ok := updates["categoryId"].(string); ok {
		t.CategoryID = &categoryIDUpdate
	}
	if date, ok := updates["date"].(string); ok {
		t.Date = date
	}
	if merchant, ok := updates["merchant"].(string); ok {
		t.Merchant = merchant
	}
	if amount, ok := updates["amount"].(float64); ok {
		t.Amount = amount
	}
	if description, ok := updates["description"].(string); ok {
		t.Description = description
	}
	if accountType, ok := updates["accountType"].(string); ok {
		t.AccountType = accountType
	}
	t.UpdatedAt = now

	_, err = tx.Exec(`
		UPDATE transactions
		SET budget_id = ?, category_id = ?, date = ?, merchant = ?, amount = ?, description = ?, account_type = ?, updated_at = ?
		WHERE id = ?
	`, t.BudgetID, t.CategoryID, t.Date, t.Merchant, t.Amount, t.Description, t.AccountType, t.UpdatedAt, id)

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &t, nil
}

func DeleteTransaction(id string) error {
	_, err := database.DB.Exec("DELETE FROM transactions WHERE id = ?", id)
	return err
}

func DeleteAllTransactions() error {
	_, err := database.DB.Exec("DELETE FROM transactions")
	return err
}

func ClearAllBudgetData() error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM transactions"); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM budgets"); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM categories"); err != nil {
		return err
	}

	return tx.Commit()
}

func ExportBudgetData() (string, error) {
	data := make(map[string]interface{})

	categories, err := GetCategories()
	if err != nil {
		return "", err
	}
	data["categories"] = categories

	budgets, err := GetBudgets()
	if err != nil {
		return "", err
	}
	data["budgets"] = budgets

	transactions, err := GetTransactions()
	if err != nil {
		return "", err
	}
	data["transactions"] = transactions

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func ImportBudgetData(jsonData string) error {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return err
	}

	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if categories, ok := data["categories"].([]interface{}); ok {
		for _, catData := range categories {
			catJSON, _ := json.Marshal(catData)
			var category Category
			if err := json.Unmarshal(catJSON, &category); err != nil {
				return err
			}
			if _, err := SaveCategory(category); err != nil {
				return err
			}
		}
	}

	if budgets, ok := data["budgets"].([]interface{}); ok {
		for _, budgetData := range budgets {
			budgetJSON, _ := json.Marshal(budgetData)
			var budget Budget
			if err := json.Unmarshal(budgetJSON, &budget); err != nil {
				return err
			}
			if _, err := SaveBudget(budget); err != nil {
				return err
			}
		}
	}

	if transactions, ok := data["transactions"].([]interface{}); ok {
		transactionsList := []Transaction{}
		for _, txData := range transactions {
			txJSON, _ := json.Marshal(txData)
			var transaction Transaction
			if err := json.Unmarshal(txJSON, &transaction); err != nil {
				return err
			}
			transactionsList = append(transactionsList, transaction)
		}
		if _, err := SaveTransactions(transactionsList); err != nil {
			return err
		}
	}

	return tx.Commit()
}

