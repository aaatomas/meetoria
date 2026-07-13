ALTER TABLE organizations
    ADD COLUMN IF NOT EXISTS currency VARCHAR(3) NOT NULL DEFAULT 'EUR';

UPDATE organizations SET currency = 'EUR' WHERE currency IS NULL OR TRIM(currency) = '';
