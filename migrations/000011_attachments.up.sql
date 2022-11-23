BEGIN;
CREATE TABLE file (
    id SERIAL PRIMARY KEY,
    company_id int REFERENCES company(id),
    backend text, -- backend can be NULL (stored in DB), or the name of a remote storage provider
    name text,
    mimetype text,
    contents bytea
);

CREATE TABLE invoice_attachments (
    invoice_id int references invoice(id),
    file_id int references file(id)
);
COMMIT;