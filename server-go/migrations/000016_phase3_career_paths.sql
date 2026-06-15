-- Phase 3: Career path definitions

CREATE TABLE IF NOT EXISTS career_paths (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    skill_requirements JSONB NOT NULL DEFAULT '[]',
    typical_progression JSONB NOT NULL DEFAULT '[]',
    industry TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS career_paths_industry_idx ON career_paths(industry);
