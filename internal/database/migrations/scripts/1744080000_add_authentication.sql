-- Migration: add_authentication
-- Created at: 2025-10-08T00:00:00Z

CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	email TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

ALTER TABLE categories ADD COLUMN user_id INTEGER REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE budgets ADD COLUMN user_id INTEGER REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE transactions ADD COLUMN user_id INTEGER REFERENCES users(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_categories_user_id ON categories(user_id);
CREATE INDEX IF NOT EXISTS idx_budgets_user_id ON budgets(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);

-- DOWN

DROP INDEX IF EXISTS idx_transactions_user_id;
DROP INDEX IF EXISTS idx_budgets_user_id;
DROP INDEX IF EXISTS idx_categories_user_id;
DROP INDEX IF EXISTS idx_users_email;

ALTER TABLE transactions DROP COLUMN user_id;
ALTER TABLE budgets DROP COLUMN user_id;
ALTER TABLE categories DROP COLUMN user_id;

DROP TABLE IF EXISTS users;

