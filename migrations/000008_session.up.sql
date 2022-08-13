BEGIN;
CREATE TABLE session (
    id text PRIMARY KEY,
    user_id int NOT NULL REFERENCES "user"(id),
    company_id int REFERENCES company(id),
    last_seen timestamp with time zone NOT NULL
);
COMMIT;