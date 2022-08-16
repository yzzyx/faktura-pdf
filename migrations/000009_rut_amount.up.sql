BEGIN;
ALTER TABLE rut_requests ADD COLUMN received_sum integer NULL;
COMMIT;