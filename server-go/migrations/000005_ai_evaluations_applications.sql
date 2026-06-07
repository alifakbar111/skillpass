-- Phase 2: AI evaluations and job applications

DO $$ BEGIN
  CREATE TYPE application_status AS ENUM ('applied', 'reviewed', 'interviewed', 'offered', 'rejected');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS ai_evaluations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID NOT NULL REFERENCES jobseeker_profiles(id) ON DELETE CASCADE,
    overall_score INTEGER NOT NULL,
    strengths JSONB NOT NULL DEFAULT '[]',
    weaknesses JSONB NOT NULL DEFAULT '[]',
    suggestions JSONB NOT NULL DEFAULT '[]',
    skill_scores JSONB NOT NULL DEFAULT '[]',
    raw_analysis TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ai_evaluations_profile_idx ON ai_evaluations(profile_id);

CREATE TABLE IF NOT EXISTS applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    jobseeker_id UUID NOT NULL REFERENCES jobseeker_profiles(id) ON DELETE CASCADE,
    job_posting_id UUID NOT NULL REFERENCES job_postings(id) ON DELETE CASCADE,
    status application_status NOT NULL DEFAULT 'applied',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(jobseeker_id, job_posting_id)
);

CREATE INDEX IF NOT EXISTS applications_jobseeker_idx ON applications(jobseeker_id);
CREATE INDEX IF NOT EXISTS applications_job_posting_idx ON applications(job_posting_id);
CREATE INDEX IF NOT EXISTS applications_status_idx ON applications(status);

-- Auto-update updated_at on row modification
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER applications_updated_at
    BEFORE UPDATE ON applications
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
