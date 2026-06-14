-- Sprint 2: SP-DID stub + working calendars

-- SP-DID records (stub, no blockchain in Phase 1)
CREATE TABLE IF NOT EXISTS sp_did_records (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id UUID NOT NULL REFERENCES companies(id),
  employee_id UUID NOT NULL UNIQUE REFERENCES employees(id),
  did_string TEXT NOT NULL UNIQUE,
  status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'revoked')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_spdid_company ON sp_did_records(company_id);
CREATE INDEX IF NOT EXISTS idx_spdid_employee ON sp_did_records(employee_id);

-- Working calendars (per-company or per-branch)
CREATE TABLE IF NOT EXISTS working_calendars (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id UUID NOT NULL REFERENCES companies(id),
  branch_id UUID REFERENCES branches(id),
  year INT NOT NULL,
  default_work_days INT[] NOT NULL DEFAULT '{1,2,3,4,5}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(company_id, branch_id, year)
);
