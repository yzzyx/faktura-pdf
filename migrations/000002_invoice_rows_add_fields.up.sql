BEGIN;
ALTER TABLE invoice_row ADD COLUMN count numeric(10,2) NOT NULL DEFAULT 1;
ALTER TABLE invoice_row ADD COLUMN vat int NOT NULL DEFAULT 0;
ALTER TABLE invoice_row ADD COLUMN unit int NOT NULL DEFAULT 0;
COMMIT;