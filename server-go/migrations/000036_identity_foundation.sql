-- server-go/migrations/000036_identity_foundation.sql
-- Phase 2 · Sprint 3: verifiable identity foundation — NO BLOCKCHAIN.
-- Ed25519 signatures + SHA-256 hashing + an append-only hash-chained log
-- deliver authenticity, integrity, and tamper-evidence without a ledger.

-- Decentralized identifiers for employees.
CREATE TABLE IF NOT EXISTS did_identities (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id     UUID NOT NULL UNIQUE REFERENCES employees(id) ON DELETE CASCADE,
    did_string      TEXT NOT NULL UNIQUE,           -- did:skillpass:<uuid>
    public_key      TEXT NOT NULL,                  -- base64 Ed25519 public key
    algorithm       TEXT NOT NULL DEFAULT 'ed25519',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at      TIMESTAMPTZ
);

-- Signed verifiable credentials (e.g. "employee of X", skill attestations).
CREATE TABLE IF NOT EXISTS signed_credentials (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    did_id            UUID NOT NULL REFERENCES did_identities(id) ON DELETE CASCADE,
    credential_type   TEXT NOT NULL,
    issuer_did        TEXT NOT NULL,
    subject_data_hash TEXT NOT NULL,                -- SHA-256 of the claim
    signature         TEXT NOT NULL,                -- base64 Ed25519 signature
    algorithm         TEXT NOT NULL DEFAULT 'ed25519',
    issued_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at        TIMESTAMPTZ,
    revoked_at        TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS signed_credentials_did_idx ON signed_credentials(did_id);

-- Append-only, hash-chained integrity log (Certificate-Transparency style).
CREATE TABLE IF NOT EXISTS integrity_anchors (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type      TEXT NOT NULL,                 -- document | credential | attendance ...
    entity_id        UUID NOT NULL,
    sha256_hash      TEXT NOT NULL,
    signature        TEXT NOT NULL,                 -- base64 signature of hash‖timestamp
    signed_by        TEXT NOT NULL,                 -- issuer DID
    prev_anchor_hash TEXT,                          -- links to the previous row
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS integrity_anchors_entity_idx ON integrity_anchors(entity_type, entity_id);
