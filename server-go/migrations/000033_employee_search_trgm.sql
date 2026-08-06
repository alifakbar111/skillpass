-- server-go/migrations/000033_employee_search_trgm.sql
-- Add pg_trgm GIN index for efficient ILIKE searches on the employees table.

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_employees_search_trgm
  ON employees
  USING gin (
    (first_name || ' ' || last_name || ' ' || email || ' ' || employee_id_number) gin_trgm_ops
  );
