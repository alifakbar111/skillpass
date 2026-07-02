-- Migration: Add years_experience_min and years_experience_max to job_postings
-- Description: Allow job postings to specify required experience range

-- Up
ALTER TABLE job_postings ADD COLUMN years_experience_min INTEGER;
ALTER TABLE job_postings ADD COLUMN years_experience_max INTEGER;

-- Down
ALTER TABLE job_postings DROP COLUMN IF EXISTS years_experience_max;
ALTER TABLE job_postings DROP COLUMN IF EXISTS years_experience_min;
