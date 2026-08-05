-- server-go/migrations/000034_face_enrollments.sql
-- Phase 2 · Sprint 2: Face recognition — enrollment + verification logs.
-- Embeddings are stored encrypted at rest (AES-256-GCM in the app layer).
-- Biometric data is subject to UU PDP consent (enforced in Sprint 8).

CREATE TABLE IF NOT EXISTS face_enrollments (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id      UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    embedding_vector BYTEA NOT NULL,             -- encrypted (AES-256-GCM)
    liveness_score   NUMERIC(4,3),
    is_active        BOOLEAN NOT NULL DEFAULT TRUE,
    enrolled_by      UUID REFERENCES users(id) ON DELETE SET NULL,
    enrolled_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS face_enrollments_employee_idx ON face_enrollments(employee_id);
-- One active enrollment per employee.
CREATE UNIQUE INDEX IF NOT EXISTS face_enrollments_active_unique
    ON face_enrollments(employee_id) WHERE is_active;

CREATE TABLE IF NOT EXISTS face_verification_logs (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id    UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    action         TEXT NOT NULL,               -- enroll | clock_in | clock_out | proctor
    match_score    NUMERIC(4,3),
    liveness_score NUMERIC(4,3),
    passed         BOOLEAN NOT NULL DEFAULT FALSE,
    ip_address     TEXT,
    user_agent     TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS face_verification_logs_employee_idx
    ON face_verification_logs(employee_id, created_at DESC);

-- ============================================================
-- PERMISSIONS + grant to existing roles
-- ============================================================
INSERT INTO permissions (id, code, module, description) VALUES
    (gen_random_uuid(), 'face.enroll', 'face', 'Enrol a face for recognition'),
    (gen_random_uuid(), 'face.view',   'face', 'View own face-enrolment status'),
    (gen_random_uuid(), 'face.admin',  'face', 'Manage face enrolments for others')
ON CONFLICT (code) DO NOTHING;

-- Employees can self-enrol and view their status; admins manage everyone.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM hris_roles r
JOIN permissions p ON (
       (p.code IN ('face.enroll', 'face.view') AND r.name IN ('Employee', 'Manager', 'Company Admin', 'HR Admin'))
    OR (p.code = 'face.admin' AND r.name IN ('Company Admin', 'HR Admin'))
)
WHERE r.is_system = true
ON CONFLICT DO NOTHING;
