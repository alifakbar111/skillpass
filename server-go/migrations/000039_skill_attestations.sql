-- server-go/migrations/000039_skill_attestations.sql
-- Phase 2 · Sprint 6: signed skill attestations + public Skill Passport.
-- Skill scores are signed with the issuer Ed25519 key (see internal/identity);
-- anyone can verify a credential against the published JWKS. NO blockchain.

-- Signed attestations of an employee's skill scores (issued after an evaluation).
CREATE TABLE IF NOT EXISTS skill_attestations (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id      UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    skill_name       TEXT NOT NULL,
    score            INTEGER NOT NULL,               -- 0..100
    evaluation_id    UUID,                           -- optional link to ai_evaluations
    attestation_hash TEXT NOT NULL,                  -- SHA-256 of the signed claim
    signature        TEXT NOT NULL,                  -- base64 Ed25519 signature
    algorithm        TEXT NOT NULL DEFAULT 'ed25519',
    issued_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at       TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS skill_attestations_employee_idx ON skill_attestations(employee_id, issued_at DESC);

-- Public Skill Passport settings — a shareable slug that renders verified badges.
CREATE TABLE IF NOT EXISTS skill_passport_public (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id     UUID NOT NULL UNIQUE REFERENCES employees(id) ON DELETE CASCADE,
    public_url_slug TEXT NOT NULL UNIQUE,
    is_public       BOOLEAN NOT NULL DEFAULT FALSE,
    settings        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS skill_passport_public_slug_idx ON skill_passport_public(public_url_slug);
