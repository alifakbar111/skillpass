-- server-go/migrations/000003_admin_audit.sql
-- Admin verification audit log.

CREATE TABLE IF NOT EXISTS admin_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    action TEXT NOT NULL,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS admin_audit_admin_idx ON admin_audit_log(admin_id);
CREATE INDEX IF NOT EXISTS admin_audit_company_idx ON admin_audit_log(company_id);
