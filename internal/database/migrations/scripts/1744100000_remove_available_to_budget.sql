-- Remove available_to_budget column from budgets table
-- This field is unnecessary - we calculate available budget from income (DEBIT) transactions
ALTER TABLE budgets DROP COLUMN available_to_budget;

-- DOWN

-- Restore available_to_budget column
ALTER TABLE budgets ADD COLUMN available_to_budget REAL NOT NULL DEFAULT 0;

