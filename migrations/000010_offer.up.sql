BEGIN;
ALTER TABLE invoice ADD COLUMN IF NOT EXISTS is_offer bool default false;
ALTER TABLE invoice ADD COLUMN IF NOT EXISTS offer_id integer REFERENCES invoice(id);
ALTER TABLE invoice ADD COLUMN IF NOT EXISTS status integer DEFAULT 0;
COMMIT;