-- Add viewer_id to profile_views (service code expects it, original migration missed it)

ALTER TABLE profile_views ADD COLUMN IF NOT EXISTS viewer_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS profile_views_viewer_idx ON profile_views(viewer_id);
