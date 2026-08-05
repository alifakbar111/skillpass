-- server-go/migrations/000037_payroll_enhanced.sql
-- Phase 2 · Sprint 4: full Indonesian payroll — PPh21 (TER) + BPJS.

-- PTKP status drives the PPh21 TER category (A/B/C).
ALTER TABLE employees ADD COLUMN IF NOT EXISTS ptkp_status TEXT NOT NULL DEFAULT 'TK/0';

-- Per-company BPJS rates + salary caps (defaults applied in-app when absent).
CREATE TABLE IF NOT EXISTS bpjs_config (
    company_id          UUID PRIMARY KEY REFERENCES companies(id) ON DELETE CASCADE,
    kesehatan_cap       NUMERIC(15,2) NOT NULL DEFAULT 12000000,
    kesehatan_employee  NUMERIC(6,5)  NOT NULL DEFAULT 0.01000,
    kesehatan_employer  NUMERIC(6,5)  NOT NULL DEFAULT 0.04000,
    jht_employee        NUMERIC(6,5)  NOT NULL DEFAULT 0.02000,
    jht_employer        NUMERIC(6,5)  NOT NULL DEFAULT 0.03700,
    jkk_employer        NUMERIC(6,5)  NOT NULL DEFAULT 0.00240,
    jkm_employer        NUMERIC(6,5)  NOT NULL DEFAULT 0.00300,
    jp_cap              NUMERIC(15,2) NOT NULL DEFAULT 10547400,
    jp_employee         NUMERIC(6,5)  NOT NULL DEFAULT 0.01000,
    jp_employer         NUMERIC(6,5)  NOT NULL DEFAULT 0.02000,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Payslip tax/BPJS breakdown columns (the JSONB breakdown carries line detail).
ALTER TABLE payslips ADD COLUMN IF NOT EXISTS bpjs_employee NUMERIC(15,2) NOT NULL DEFAULT 0;
ALTER TABLE payslips ADD COLUMN IF NOT EXISTS bpjs_employer NUMERIC(15,2) NOT NULL DEFAULT 0;
ALTER TABLE payslips ADD COLUMN IF NOT EXISTS pph21         NUMERIC(15,2) NOT NULL DEFAULT 0;

-- Run-level roll-ups.
ALTER TABLE payroll_runs ADD COLUMN IF NOT EXISTS total_bpjs_employee NUMERIC(15,2) NOT NULL DEFAULT 0;
ALTER TABLE payroll_runs ADD COLUMN IF NOT EXISTS total_bpjs_employer NUMERIC(15,2) NOT NULL DEFAULT 0;
ALTER TABLE payroll_runs ADD COLUMN IF NOT EXISTS total_pph21         NUMERIC(15,2) NOT NULL DEFAULT 0;

-- Bank-transfer records generated per employee per run (for disbursement export).
CREATE TABLE IF NOT EXISTS bank_transfers (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payroll_run_id UUID NOT NULL REFERENCES payroll_runs(id) ON DELETE CASCADE,
    employee_id    UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    bank_name      TEXT,
    account_number TEXT,
    account_holder TEXT,
    amount         NUMERIC(15,2) NOT NULL DEFAULT 0,
    status         TEXT NOT NULL DEFAULT 'pending',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS bank_transfers_run_idx ON bank_transfers(payroll_run_id);
