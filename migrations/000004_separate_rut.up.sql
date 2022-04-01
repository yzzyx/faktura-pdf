BEGIN;
CREATE TABLE rut_requests (
    id SERIAL PRIMARY KEY,
    type int NOT NULL DEFAULT 0,
    invoice_id int NOT NULL,
    status int NOT NULL DEFAULT 0,
    date_sent timestamp with time zone,
    date_paid timestamp with time zone
);

INSERT INTO rut_requests (invoice_id, status)
SELECT id, 2 FROM invoice WHERE is_rut_paid;

INSERT INTO rut_requests (invoice_id, status)
SELECT id, 1 FROM invoice WHERE is_rut_sent AND not is_rut_paid;

ALTER TABLE invoice DROP COLUMN is_rut_sent;
ALTER TABLE invoice DROP COLUMN is_rut_paid;
COMMIT;