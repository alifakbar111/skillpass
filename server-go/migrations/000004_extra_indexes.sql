-- server-go/migrations/000004_extra_indexes.sql
-- Performance indexes for list queries and FK lookups.

CREATE INDEX IF NOT EXISTS companies_verification_status_idx ON companies(verification_status, created_at);
CREATE INDEX IF NOT EXISTS jobseeker_profiles_user_id_idx ON jobseeker_profiles(user_id);
