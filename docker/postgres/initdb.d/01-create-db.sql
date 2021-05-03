CREATE SCHEMA ext;
GRANT ALL PRIVILEGES ON SCHEMA ext TO faktura;
-- Must be added in its own schema to avoid flyway complaining on unknown tables created by extension
CREATE EXTENSION pg_stat_statements SCHEMA ext;
CREATE EXTENSION unaccent;
