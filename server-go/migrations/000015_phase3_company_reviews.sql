-- Phase 3: Candidate reviews of companies

CREATE TYPE interaction_type AS ENUM ('applied', 'interviewed');

CREATE TABLE IF NOT EXISTS company_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    candidate_id UUID NOT NULL REFERENCES jobseeker_profiles(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    review TEXT,
    interaction_type interaction_type NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(company_id, candidate_id)
);

CREATE INDEX IF NOT EXISTS company_reviews_company_idx ON company_reviews(company_id);
CREATE INDEX IF NOT EXISTS company_reviews_candidate_idx ON company_reviews(candidate_id);
