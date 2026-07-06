-- ============================================================
-- Migration 000026: Activity Logs (HRIS audit trail)
-- ============================================================

CREATE TABLE IF NOT EXISTS activity_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    actor_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS activity_logs_company_created_idx
    ON activity_logs(company_id, created_at DESC);

CREATE INDEX IF NOT EXISTS activity_logs_entity_idx
    ON activity_logs(entity_type, entity_id)
    WHERE entity_id IS NOT NULL;

-- Add permission for activity log
INSERT INTO permissions (id, code, module, description)
VALUES (gen_random_uuid(), 'activity.view', 'org', 'View activity log')
ON CONFLICT (code) DO NOTHING;

-- Grant activity.view to roles that already have org.view
INSERT INTO role_permissions (role_id, permission_id)
SELECT rp.role_id, p.id
FROM role_permissions rp
CROSS JOIN permissions p
WHERE rp.permission_id = (SELECT id FROM permissions WHERE code = 'org.view')
  AND p.code = 'activity.view'
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp2
    WHERE rp2.role_id = rp.role_id AND rp2.permission_id = p.id
  )
ON CONFLICT DO NOTHING;
