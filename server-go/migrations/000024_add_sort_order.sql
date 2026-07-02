ALTER TABLE job_experiences ADD COLUMN IF NOT EXISTS sort_order INTEGER NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_job_experiences_sort ON job_experiences(profile_id, sort_order);
