BEGIN;
CREATE TABLE company (
    id SERIAL PRIMARY KEY,
    name text NOT NULL default '',
    email text NOT NULL default '',
    address1 text NOT NULL default '',
    address2 text NOT NULL default '',
    postcode text NOT NULL default '',
    city text NOT NULL default '',
    telephone text NOT NULL default '',
    company_id text NOT NULL default '',
    payment_account text NOT NULL default '',
    payment_type int NOT NULL default 1,
    vat_number text NOT NULL default '',

    invoice_number int NOT NULL default 1,
    invoice_due_days int NOT NULL default 30,
    invoice_reference text NOT NULL default '',
    invoice_text text NOT NULL default '',
    invoice_template text NOT NULL default '',

    offer_text text NOT NULL default '',
    offer_template text NOT NULL default ''
);

ALTER TABLE customer ADD COLUMN company_id int REFERENCES company(id);
ALTER TABLE invoice ADD COLUMN company_id int REFERENCES company(id);
ALTER TABLE rut_requests ADD COLUMN company_id int REFERENCES company(id);

CREATE TABLE company_user (
    user_id int REFERENCES "user"(id),
    company_id int REFERENCES company(id)
);
COMMIT;