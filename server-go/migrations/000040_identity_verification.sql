-- server-go/migrations/000040_identity_verification.sql
-- Phase 2 · Sprint 6: external identity verification (Dukcapil KTP/NIK, PDDikti
-- education). Providers are stub-able — this table records the outcome per
-- employee regardless of which provider (or a manual check) produced it.

DO $$ BEGIN
  CREATE TYPE identity_provider AS ENUM ('dukcapil', 'pddikti', 'manual');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
  CREATE TYPE identity_verify_status AS ENUM ('pending', 'verified', 'failed', 'unverified');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS identity_verifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    provider        identity_provider NOT NULL,
    response_status identity_verify_status NOT NULL DEFAULT 'pending',
    detail          TEXT,                            -- human-readable note / provider message
    verified_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS identity_verifications_employee_idx ON identity_verifications(employee_id, created_at DESC);
