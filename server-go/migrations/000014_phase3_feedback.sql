-- Phase 3: Company-to-candidate feedback

CREATE TABLE IF NOT EXISTS feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID NOT NULL REFERENCES jobseeker_profiles(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    content TEXT NOT NULL DEFAULT '',
    rating_areas JSONB NOT NULL DEFAULT '[]',
    ai_suggestions JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS feedback_profile_idx ON feedback(profile_id);
CREATE INDEX IF NOT EXISTS feedback_company_idx ON feedback(company_id);
