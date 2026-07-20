-- token_version claim (PR-20 of the 2026-07-20 hardening roadmap).
--
-- A user whose role was just downgraded by an admin still has a valid
-- 15-minute access token issued with the old role claim. Bumping
-- token_version on role change invalidates outstanding tokens without
-- rotating JWT_SECRET.
--
-- Initial value 0 (matching pre-existing rows). Auth middleware
-- compares the JWT claim to the DB column; on mismatch, returns 401.

ALTER TABLE users ADD COLUMN IF NOT EXISTS token_version INTEGER NOT NULL DEFAULT 0;
