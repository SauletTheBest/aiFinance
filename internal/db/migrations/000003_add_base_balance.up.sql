-- Add base_balance to users to support the "Opening Balance" pattern.
-- base_balance represents the user's account balance before their first tracked transaction.
-- Actual balance = base_balance + (total income - total expenses from transactions).
ALTER TABLE users ADD COLUMN IF NOT EXISTS base_balance DOUBLE PRECISION NOT NULL DEFAULT 0;
