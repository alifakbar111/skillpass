-- server-go/migrations/000038_ats_full.sql
-- Phase 2 · Sprint 5: ATS full pipeline + ATS→HRIS account bridge.
-- Configurable hiring pipelines, per-stage scorecards, interview scheduling,
-- and template-based offer letters with in-app e-signature. On offer
-- acceptance the candidate's EXISTING login is linked to a new employee record
-- (see internal/hris/ats/onboard.go) — one identity across jobseeker + HRIS.

-- ============================================================
-- ENUMS
-- ============================================================
DO $$ BEGIN
  CREATE TYPE ats_stage_type AS ENUM
    ('screening', 'phone_screen', 'technical', 'hr_interview', 'final', 'offer', 'hired');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
  CREATE TYPE ats_candidate_status AS ENUM ('active', 'hired', 'rejected', 'withdrawn');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
  CREATE TYPE ats_offer_status AS ENUM ('draft', 'sent', 'accepted', 'declined', 'expired');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
  CREATE TYPE ats_interview_status AS ENUM ('scheduled', 'completed', 'cancelled', 'no_show');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

-- ============================================================
-- PIPELINES + STAGES (configurable per company)
-- ============================================================
CREATE TABLE IF NOT EXISTS ats_pipelines (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id  UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    is_default  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ats_pipelines_company_idx ON ats_pipelines(company_id);

CREATE TABLE IF NOT EXISTS ats_pipeline_stages (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pipeline_id UUID NOT NULL REFERENCES ats_pipelines(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    stage_type  ats_stage_type NOT NULL DEFAULT 'screening',
    sort_order  INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ats_pipeline_stages_pipeline_idx ON ats_pipeline_stages(pipeline_id, sort_order);

-- ============================================================
-- CANDIDATES (bridges to the existing applications table)
-- ============================================================
CREATE TABLE IF NOT EXISTS ats_candidates (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id       UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    pipeline_id      UUID NOT NULL REFERENCES ats_pipelines(id) ON DELETE CASCADE,
    current_stage_id UUID REFERENCES ats_pipeline_stages(id) ON DELETE SET NULL,
    application_id   UUID REFERENCES applications(id) ON DELETE SET NULL,
    jobseeker_id     UUID REFERENCES jobseeker_profiles(id) ON DELETE SET NULL,
    job_posting_id   UUID REFERENCES job_postings(id) ON DELETE SET NULL,
    candidate_name   TEXT NOT NULL,
    candidate_email  TEXT NOT NULL,
    status           ats_candidate_status NOT NULL DEFAULT 'active',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ats_candidates_company_idx  ON ats_candidates(company_id);
CREATE INDEX IF NOT EXISTS ats_candidates_pipeline_idx ON ats_candidates(pipeline_id);
CREATE INDEX IF NOT EXISTS ats_candidates_stage_idx    ON ats_candidates(current_stage_id);
-- One ATS candidate per application (a job application enters the pipeline once).
CREATE UNIQUE INDEX IF NOT EXISTS ats_candidates_application_uidx
    ON ats_candidates(application_id) WHERE application_id IS NOT NULL;

-- ============================================================
-- SCORECARDS (per-evaluator structured feedback)
-- ============================================================
CREATE TABLE IF NOT EXISTS ats_scorecards (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    candidate_id   UUID NOT NULL REFERENCES ats_candidates(id) ON DELETE CASCADE,
    stage_id       UUID REFERENCES ats_pipeline_stages(id) ON DELETE SET NULL,
    evaluator_id   UUID REFERENCES users(id) ON DELETE SET NULL,
    evaluator_name TEXT NOT NULL DEFAULT '',
    scores         JSONB NOT NULL DEFAULT '{}'::jsonb,
    overall_rating INTEGER,                              -- 1..5
    recommendation TEXT,                                 -- strong_yes | yes | no | strong_no
    notes          TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ats_scorecards_candidate_idx ON ats_scorecards(candidate_id, created_at DESC);

-- ============================================================
-- INTERVIEWS (scheduling; standalone so a candidate need not have an application)
-- ============================================================
CREATE TABLE IF NOT EXISTS ats_interviews (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    candidate_id  UUID NOT NULL REFERENCES ats_candidates(id) ON DELETE CASCADE,
    stage_id      UUID REFERENCES ats_pipeline_stages(id) ON DELETE SET NULL,
    scheduled_at  TIMESTAMPTZ NOT NULL,
    mode          TEXT NOT NULL DEFAULT 'onsite',        -- onsite | online
    location      TEXT,
    meeting_link  TEXT,
    interviewer   TEXT,
    notes         TEXT,
    status        ats_interview_status NOT NULL DEFAULT 'scheduled',
    created_by    UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ats_interviews_candidate_idx ON ats_interviews(candidate_id, scheduled_at DESC);

-- ============================================================
-- OFFER TEMPLATES + OFFER LETTERS (in-app e-signature)
-- ============================================================
CREATE TABLE IF NOT EXISTS ats_offer_templates (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    body       TEXT NOT NULL,                            -- merge fields: {{candidateName}}, {{position}}, {{salary}}, {{startDate}}
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ats_offer_templates_company_idx ON ats_offer_templates(company_id);

CREATE TABLE IF NOT EXISTS ats_offer_letters (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    candidate_id   UUID NOT NULL REFERENCES ats_candidates(id) ON DELETE CASCADE,
    template_id    UUID REFERENCES ats_offer_templates(id) ON DELETE SET NULL,
    position_title TEXT NOT NULL,
    salary         NUMERIC(15,2),
    start_date     DATE,
    body           TEXT NOT NULL,                        -- rendered body (merge fields resolved)
    status         ats_offer_status NOT NULL DEFAULT 'draft',
    accept_token   TEXT NOT NULL UNIQUE,                 -- opaque token for the public accept link
    signature_name TEXT,                                 -- typed signature captured on acceptance
    signed_at      TIMESTAMPTZ,
    sent_at        TIMESTAMPTZ,
    created_by     UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ats_offer_letters_candidate_idx ON ats_offer_letters(candidate_id, created_at DESC);

-- ============================================================
-- PERMISSIONS: add ats.interview + ats.offer (ats.view/manage/scorecard exist)
-- ============================================================
INSERT INTO permissions (id, code, module, description) VALUES
    (gen_random_uuid(), 'ats.interview', 'ats', 'Schedule and manage candidate interviews'),
    (gen_random_uuid(), 'ats.offer',     'ats', 'Generate and send offer letters')
ON CONFLICT (code) DO NOTHING;

-- Grant to existing system roles so already-seeded companies get access without
-- re-seeding. New companies pick these up via rbac.seedRolePermissions.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM hris_roles r
JOIN permissions p ON p.code IN ('ats.interview', 'ats.offer')
WHERE r.is_system = true AND r.name IN ('Company Admin', 'HR Admin', 'Recruiter')
ON CONFLICT DO NOTHING;

-- Managers can schedule interviews (they already hold ats.scorecard).
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM hris_roles r
JOIN permissions p ON p.code = 'ats.interview'
WHERE r.is_system = true AND r.name = 'Manager'
ON CONFLICT DO NOTHING;
