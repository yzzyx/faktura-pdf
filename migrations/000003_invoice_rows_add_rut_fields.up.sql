BEGIN;
ALTER TABLE invoice_row ADD COLUMN rot_rut_service_type int;
COMMIT;