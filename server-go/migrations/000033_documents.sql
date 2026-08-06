-- server-go/migrations/000033_documents.sql
-- Phase 2 · Sprint 1: Document Management.
-- Company-scoped document store with SHA-256 integrity, malware-scan status,
-- and an access audit log. Files live in private storage (S3 or a private
-- local dir), never the public /uploads path.

-- ============================================================
-- ENUMS
-- ============================================================
DO $$ BEGIN
  CREATE TYPE document_category AS ENUM ('identity', 'contract', 'certificate', 'payslip', 'tax', 'other');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
  CREATE TYPE scan_status AS ENUM ('pending', 'clean', 'infected', 'error');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
  CREATE TYPE export_status AS ENUM ('queued', 'processing', 'done', 'failed');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

-- ============================================================
-- TABLES
-- ============================================================
CREATE TABLE IF NOT EXISTS documents (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id        UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    employee_id       UUID REFERENCES employees(id) ON DELETE SET NULL,
    category          document_category NOT NULL DEFAULT 'other',
    original_filename TEXT NOT NULL,
    storage_key       TEXT NOT NULL,
    sha256_hash       TEXT NOT NULL,
    mime_type         TEXT NOT NULL,
    file_size         BIGINT NOT NULL,
    scan_status       scan_status NOT NULL DEFAULT 'pending',
    uploaded_by       UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS documents_company_idx  ON documents(company_id);
CREATE INDEX IF NOT EXISTS documents_employee_idx ON documents(employee_id);

CREATE TABLE IF NOT EXISTS document_access_log (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    accessed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    action      TEXT NOT NULL,            -- upload | download | delete | view
    ip_address  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS document_access_log_doc_idx ON document_access_log(document_id, created_at DESC);

CREATE TABLE IF NOT EXISTS export_jobs (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id   UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    type         TEXT NOT NULL,
    status       export_status NOT NULL DEFAULT 'queued',
    file_url     TEXT,
    requested_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS export_jobs_company_idx ON export_jobs(company_id, created_at DESC);

-- ============================================================
-- PERMISSIONS (global) + grant to existing system roles
-- ============================================================
INSERT INTO permissions (id, code, module, description) VALUES
    (gen_random_uuid(), 'documents.upload',    'documents', 'Upload documents'),
    (gen_random_uuid(), 'documents.delete',    'documents', 'Delete documents'),
    (gen_random_uuid(), 'documents.audit_log', 'documents', 'View document access audit log')
ON CONFLICT (code) DO NOTHING;

-- Grant the new permissions to every existing Company Admin / HR Admin role so
-- already-seeded companies get access without re-seeding. (New companies pick
-- them up via rbac.seedRolePermissions.)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM hris_roles r
JOIN permissions p ON p.code IN ('documents.upload', 'documents.delete', 'documents.audit_log')
WHERE r.is_system = true AND r.name IN ('Company Admin', 'HR Admin')
ON CONFLICT DO NOTHING;
