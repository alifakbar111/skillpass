-- HIGH-005: ai_evaluations.raw_analysis previously stored the full LLM
-- system prompt + user PII (profile name, experience entries with
-- descriptions and skills). A DB leak exposed competitive intelligence
-- and personal data. The column is not read by any handler.
--
-- Strategy: keep the column for backwards compatibility, but
-- 1) redact every existing row to an empty string, and
-- 2) the service layer is being changed to never write PII again.
-- Future writes will store a short fingerprint of skills for debugging
-- without leaking PII.

UPDATE ai_evaluations
SET raw_analysis = ''
WHERE raw_analysis IS NOT NULL AND raw_analysis <> '';
