-- ============================================================
-- Migration 000027: Evaluation Lifecycle (is_current flag)
-- ============================================================

-- Add is_current flag for evaluation history tracking
-- Default true for backward compatibility with existing evaluations
ALTER TABLE ai_evaluations ADD COLUMN IF NOT EXISTS is_current BOOLEAN NOT NULL DEFAULT true;

-- Partial index for fast "current evaluation per profile" lookups
CREATE INDEX IF NOT EXISTS idx_ai_evaluations_current_profile
  ON ai_evaluations(profile_id, is_current)
  WHERE is_current = true;

-- Index for history queries (all evaluations for a profile, ordered by time)
CREATE INDEX IF NOT EXISTS idx_ai_evaluations_profile_created
  ON ai_evaluations(profile_id, created_at DESC);
