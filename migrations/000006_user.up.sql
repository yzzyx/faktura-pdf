BEGIN;
CREATE TABLE "user" (
    id SERIAL PRIMARY KEY,
    username varchar(50) NOT NULL,
    password text NOT NULL,

    name text NOT NULL,
    email text NOT NULL
);
COMMIT;