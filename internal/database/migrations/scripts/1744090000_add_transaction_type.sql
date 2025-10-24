-- Add transaction_type column to transactions table
ALTER TABLE transactions ADD COLUMN transaction_type TEXT NOT NULL DEFAULT 'DEBIT';

-- Create index for efficient queries by transaction type
CREATE INDEX idx_transactions_type ON transactions(transaction_type);

-- Update existing transactions to be DEBIT (expenses) by default
UPDATE transactions SET transaction_type = 'DEBIT' WHERE transaction_type IS NULL OR transaction_type = '';

-- DOWN

-- Remove index
DROP INDEX IF EXISTS idx_transactions_type;

-- Remove transaction_type column
ALTER TABLE transactions DROP COLUMN transaction_type;

