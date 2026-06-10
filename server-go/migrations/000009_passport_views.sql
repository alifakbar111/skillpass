-- Phase 2: Passport view counter

ALTER TABLE jobseeker_profiles ADD COLUMN IF NOT EXISTS view_count INTEGER NOT NULL DEFAULT 0;
