CREATE TABLE IF NOT EXISTS customer (
   id SERIAL PRIMARY KEY,
   name text NOT NULL DEFAULT '',
   email text NOT NULL DEFAULT '',
   address1 text NOT NULL DEFAULT '',
   address2 text NOT NULL DEFAULT '',
   postcode text NOT NULL DEFAULT '',
   city text NOT NULL DEFAULT '',
   pnr text NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS invoice (
   id SERIAL PRIMARY KEY,
   number int UNIQUE,
   date_created timestamp with time zone NOT NULL DEFAULT current_timestamp,
   date_invoiced timestamp with time zone,
   date_due timestamp with time zone,
   date_payed timestamp with time zone,
   name text NOT NULL,
   customer_id int NOT NULL,
   is_offered boolean NOT NULL default false,
   is_invoiced boolean NOT NULL default false,
   is_payed boolean NOT NULL default false,
   is_deleted boolean NOT NULL default false,
   additional_info text NOT NULL default '',
   FOREIGN KEY(customer_id)
       REFERENCES customer(id)
);

CREATE TABLE IF NOT EXISTS invoice_row (
    id SERIAL PRIMARY KEY,
    row_order int NOT NULL DEFAULT 0,
    invoice_id int NOT NULL,
    description text NOT NULL,
    cost numeric(10,2) NOT NULL,
    is_rot_rut boolean,
    FOREIGN KEY(invoice_id)
        REFERENCES invoice(id)
);