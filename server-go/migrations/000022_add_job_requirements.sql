-- Add requirements column to job_postings table

ALTER TABLE job_postings ADD COLUMN IF NOT EXISTS requirements TEXT;
