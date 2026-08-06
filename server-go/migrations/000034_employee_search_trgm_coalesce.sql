-- server-go/migrations/000034_employee_search_trgm_coalesce.sql
-- Rebuild the employee search index from 000033 with COALESCE.
--
-- The old index expression (first_name || ' ' || last_name || ...) yields NULL
-- for any row with a NULL first/last name/email, silently dropping those rows
-- from the index so they can never match a search. COALESCE keeps them in.

DROP INDEX IF EXISTS idx_employees_search_trgm;

CREATE INDEX IF NOT EXISTS idx_employees_search_trgm
  ON employees
  USING gin (
    (COALESCE(first_name, '') || ' ' || COALESCE(last_name, '') || ' ' || COALESCE(email, '') || ' ' || COALESCE(employee_id_number, '')) gin_trgm_ops
  );
