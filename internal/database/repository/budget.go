package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"api.alexmontague.ca/internal/database"
	"github.com/google/uuid"
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
	ID          string             `json:"id"`
	Month       string             `json:"month"`
	Allocations map[string]float64 `json:"allocations"`
	CreatedAt   string             `json:"createdAt"`
	UpdatedAt   string             `json:"updatedAt"`
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
	TransactionType string  `json:"transactionType"` // "DEBIT" (income) or "CREDIT" (expense)
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

func GetCategories(userID int) ([]Category, error) {
	rows, err := database.DB.Query(`
		SELECT id, name, monthly_budget, color, created_at, updated_at
		FROM categories
		WHERE user_id = ?
		ORDER BY name
	`, userID)
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

func SaveCategory(category Category, userID int) (*Category, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	category.ID = uuid.New().String()
	category.CreatedAt = now
	category.UpdatedAt = now

	_, err := database.DB.Exec(`
		INSERT INTO categories (id, name, monthly_budget, color, user_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, category.ID, category.Name, category.MonthlyBudget, category.Color, userID, category.CreatedAt, category.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &category, nil
}

func UpdateCategory(id string, userID int, updates map[string]interface{}) (*Category, error) {
	now := time.Now().Format("2006-01-02 15:04:05")

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var category Category
	err = tx.QueryRow(`
		SELECT id, name, monthly_budget, color, created_at, updated_at
		FROM categories WHERE id = ? AND user_id = ?
	`, id, userID).Scan(&category.ID, &category.Name, &category.MonthlyBudget, &category.Color, &category.CreatedAt, &category.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if name, ok := updates["name"].(string); ok {
		if category.Name != "Other" {
			category.Name = name
		}
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

func DeleteCategory(id string, userID int) error {
	var categoryName string
	err := database.DB.QueryRow("SELECT name FROM categories WHERE id = ? AND user_id = ?", id, userID).Scan(&categoryName)
	if err != nil {
		return err
	}

	if categoryName == "Other" {
		return sql.ErrNoRows
	}

	_, err = database.DB.Exec("DELETE FROM categories WHERE id = ? AND user_id = ?", id, userID)
	return err
}

func DeleteAllCategories(userID int) error {
	_, err := database.DB.Exec("DELETE FROM categories WHERE user_id = ? AND name != 'Other'", userID)
	return err
}

func CreateDefaultCategories(userID int) error {
	defaultCategories := []struct {
		name  string
		color string
	}{
		{"Groceries", "#22c55e"},
		{"Dining Out", "#ef4444"},
		{"Entertainment", "#ec4899"},
		{"Shopping", "#8b5cf6"},
		{"Transportation", "#3b82f6"},
		{"Bills & Utilities", "#f59e0b"},
		{"Subscriptions", "#06b6d4"},
		{"Mortgage", "#06b6d4"},
		{"Insurance", "#06b6d4"},
		{"Healthcare", "#10b981"},
		{"Other", "#6b7280"},
	}

	for _, dc := range defaultCategories {
		color := dc.color
		category := Category{
			Name:  dc.name,
			Color: &color,
		}
		_, err := SaveCategory(category, userID)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetBudgets(userID int) ([]Budget, error) {
	rows, err := database.DB.Query(`
		SELECT id, month, allocations, created_at, updated_at
		FROM budgets
		WHERE user_id = ?
		ORDER BY month DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	budgets := []Budget{}
	for rows.Next() {
		var b Budget
		var allocationsJSON string
		err := rows.Scan(&b.ID, &b.Month, &allocationsJSON, &b.CreatedAt, &b.UpdatedAt)
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

func SaveBudget(budget Budget, userID int) (*Budget, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	budget.ID = uuid.New().String()
	budget.CreatedAt = now
	budget.UpdatedAt = now

	allocationsJSON, err := json.Marshal(budget.Allocations)
	if err != nil {
		return nil, err
	}

	_, err = database.DB.Exec(`
		INSERT INTO budgets (id, month, allocations, user_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, budget.ID, budget.Month, string(allocationsJSON), userID, budget.CreatedAt, budget.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &budget, nil
}

func UpdateBudget(id string, userID int, updates map[string]interface{}) (*Budget, error) {
	now := time.Now().Format("2006-01-02 15:04:05")

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var budget Budget
	var allocationsJSON string
	err = tx.QueryRow(`
		SELECT id, month, allocations, created_at, updated_at
		FROM budgets WHERE id = ? AND user_id = ?
	`, id, userID).Scan(&budget.ID, &budget.Month, &allocationsJSON, &budget.CreatedAt, &budget.UpdatedAt)
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
	budget.UpdatedAt = now

	newAllocationsJSON, err := json.Marshal(budget.Allocations)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`
		UPDATE budgets
		SET month = ?, allocations = ?, updated_at = ?
		WHERE id = ?
	`, budget.Month, string(newAllocationsJSON), budget.UpdatedAt, id)

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &budget, nil
}

func DeleteBudget(id string, userID int) error {
	_, err := database.DB.Exec("DELETE FROM budgets WHERE id = ? AND user_id = ?", id, userID)
	return err
}

func DeleteAllBudgets(userID int) error {
	_, err := database.DB.Exec("DELETE FROM budgets WHERE user_id = ?", userID)
	return err
}

func GetTransactions(userID int) ([]Transaction, error) {
	rows, err := database.DB.Query(`
		SELECT id, budget_id, category_id, transaction_hash, date, merchant, amount, description, account_type, transaction_type, created_at, updated_at
		FROM transactions
		WHERE user_id = ?
		ORDER BY date DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := []Transaction{}
	for rows.Next() {
		var t Transaction
		var categoryID sql.NullString
		err := rows.Scan(&t.ID, &t.BudgetID, &categoryID, &t.TransactionHash, &t.Date, &t.Merchant, &t.Amount, &t.Description, &t.AccountType, &t.TransactionType, &t.CreatedAt, &t.UpdatedAt)
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

func SaveTransactions(transactions []Transaction, userID int) ([]Transaction, error) {
	categories, err := GetCategories(userID)
	if err != nil {
		return nil, err
	}

	budgets, err := GetBudgets(userID)
	if err != nil {
		return nil, err
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	transactionsToSave := make([]Transaction, len(transactions))
	budgetCache := make(map[string]string)

	for _, b := range budgets {
		budgetCache[b.Month] = b.ID
	}

	for i, t := range transactions {
		t.ID = uuid.New().String()
		t.CreatedAt = now
		t.UpdatedAt = now

		if t.TransactionType == "" {
			t.TransactionType = "CREDIT"
		}

		transactionMonth := t.Date[:7]
		if budgetID, exists := budgetCache[transactionMonth]; exists {
			t.BudgetID = budgetID
		} else {
			allocations := make(map[string]float64)
			for _, cat := range categories {
				if cat.MonthlyBudget != nil {
					allocations[cat.ID] = *cat.MonthlyBudget
				} else {
					allocations[cat.ID] = 0
				}
			}

			newBudget := Budget{
				Month:       transactionMonth,
				Allocations: allocations,
			}

			savedBudget, err := SaveBudget(newBudget, userID)
			if err != nil {
				return nil, err
			}

			budgetCache[transactionMonth] = savedBudget.ID
			t.BudgetID = savedBudget.ID
		}

		transactionsToSave[i] = t
	}

	if len(categories) > 0 && len(transactionsToSave) > 0 {
		categorized, err := CategorizeTransactions(categories, transactionsToSave)
		if err == nil && len(categorized) > 0 {
			transactionsToSave = ApplyCategorizationToTransactions(categories, categorized, transactionsToSave)

			fmt.Println("\n=== AI CATEGORIZATION RESULTS ===")
			categoryMap := make(map[string]string)
			for _, cat := range categories {
				categoryMap[cat.ID] = cat.Name
			}

			for _, tx := range transactionsToSave {
				categoryName := "NONE"
				if tx.CategoryID != nil {
					if name, exists := categoryMap[*tx.CategoryID]; exists {
						categoryName = name
					}
				}
				fmt.Printf("Transaction: %-40s | Merchant: %-30s | Amount: $%-8.2f | Category: %s\n",
					tx.Description, tx.Merchant, tx.Amount, categoryName)
			}
			fmt.Println("================================\n")
		} else if err != nil {
			fmt.Printf("AI Categorization Error: %v\n", err)
		}
	}

	dbTx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer dbTx.Rollback()

	stmt, err := dbTx.Prepare(`
		INSERT INTO transactions (id, budget_id, category_id, transaction_hash, date, merchant, amount, description, account_type, transaction_type, user_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	savedTransactions := []Transaction{}
	for _, t := range transactionsToSave {
		_, err = stmt.Exec(t.ID, t.BudgetID, t.CategoryID, t.TransactionHash, t.Date, t.Merchant, t.Amount, t.Description, t.AccountType, t.TransactionType, userID, t.CreatedAt, t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		savedTransactions = append(savedTransactions, t)
	}

	if err := dbTx.Commit(); err != nil {
		return nil, err
	}

	return savedTransactions, nil
}

func UpdateTransaction(id string, userID int, updates map[string]interface{}) (*Transaction, error) {
	now := time.Now().Format("2006-01-02 15:04:05")

	tx, err := database.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var t Transaction
	var categoryID sql.NullString
	err = tx.QueryRow(`
		SELECT id, budget_id, category_id, transaction_hash, date, merchant, amount, description, account_type, transaction_type, created_at, updated_at
		FROM transactions WHERE id = ? AND user_id = ?
	`, id, userID).Scan(&t.ID, &t.BudgetID, &categoryID, &t.TransactionHash, &t.Date, &t.Merchant, &t.Amount, &t.Description, &t.AccountType, &t.TransactionType, &t.CreatedAt, &t.UpdatedAt)
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
	if transactionType, ok := updates["transactionType"].(string); ok {
		t.TransactionType = transactionType
	}
	t.UpdatedAt = now

	_, err = tx.Exec(`
		UPDATE transactions
		SET budget_id = ?, category_id = ?, date = ?, merchant = ?, amount = ?, description = ?, account_type = ?, transaction_type = ?, updated_at = ?
		WHERE id = ?
	`, t.BudgetID, t.CategoryID, t.Date, t.Merchant, t.Amount, t.Description, t.AccountType, t.TransactionType, t.UpdatedAt, id)

	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &t, nil
}

func DeleteTransaction(id string, userID int) error {
	_, err := database.DB.Exec("DELETE FROM transactions WHERE id = ? AND user_id = ?", id, userID)
	return err
}

func DeleteAllTransactions(userID int) error {
	_, err := database.DB.Exec("DELETE FROM transactions WHERE user_id = ?", userID)
	return err
}

func ClearAllBudgetData(userID int) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM transactions WHERE user_id = ?", userID); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM budgets WHERE user_id = ?", userID); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM categories WHERE user_id = ?", userID); err != nil {
		return err
	}

	return tx.Commit()
}

func ExportBudgetData(userID int) (string, error) {
	data := make(map[string]interface{})

	categories, err := GetCategories(userID)
	if err != nil {
		return "", err
	}
	data["categories"] = categories

	budgets, err := GetBudgets(userID)
	if err != nil {
		return "", err
	}
	data["budgets"] = budgets

	transactions, err := GetTransactions(userID)
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

func ImportBudgetData(jsonData string, userID int) error {
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
			if _, err := SaveCategory(category, userID); err != nil {
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
			if _, err := SaveBudget(budget, userID); err != nil {
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
		if _, err := SaveTransactions(transactionsList, userID); err != nil {
			return err
		}
	}

	return tx.Commit()
}
