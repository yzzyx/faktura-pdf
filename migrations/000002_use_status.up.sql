CREATE TABLE IF NOT EXISTS status (
    id SERIAL PRIMARY KEY,
    name text NOT NULL,
    description text,
    class text
);

ALTER TABLE invoice ADD COLUMN status_id int;
ALTER TABLE invoice ADD CONSTRAINT invoice_fk_status_id FOREIGN KEY(status_id) REFERENCES status(id);

INSERT INTO status (name, description, class) VALUES
                                 ('Offert skickad', '', 'badge badge-primary'),
                                 ('Faktura skickad', '', 'badge badge-warning'),
                                 ('Faktura betald', '', 'badge badge-success'),
                                 ('Påminnelse skickad', '', 'badge badge-error'),
                                 ('RUT skickad', '', 'badge badge-warning'),
                                 ('RUT betald', '', 'badge badge-success'),
                                 ('Klar', '', 'badge badge-success');
