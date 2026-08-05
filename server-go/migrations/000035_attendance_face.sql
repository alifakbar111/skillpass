-- server-go/migrations/000035_attendance_face.sql
-- Phase 2 · Sprint 3: link face verification into attendance clock-in/out.

ALTER TABLE attendance_logs ADD COLUMN IF NOT EXISTS face_verification_id UUID;
ALTER TABLE attendance_logs ADD COLUMN IF NOT EXISTS face_match_score NUMERIC(4,3);
