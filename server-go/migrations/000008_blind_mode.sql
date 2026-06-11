-- Phase 2: Blind hiring mode (company setting to mask candidate identities during screening)

ALTER TABLE companies ADD COLUMN IF NOT EXISTS blind_mode BOOLEAN NOT NULL DEFAULT FALSE;
