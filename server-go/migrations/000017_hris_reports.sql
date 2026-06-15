-- Sprint 6: HR Analytics & Reporting
CREATE TABLE IF NOT EXISTS hr_analytics_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id),
    snapshot_month DATE NOT NULL,
    total_headcount INT NOT NULL DEFAULT 0,
    new_hires INT NOT NULL DEFAULT 0,
    terminations INT NOT NULL DEFAULT 0,
    turnover_rate NUMERIC(5,2) NOT NULL DEFAULT 0,
    avg_tenure_months NUMERIC(7,2) NOT NULL DEFAULT 0,
    department_breakdown JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(company_id, snapshot_month)
);

CREATE INDEX IF NOT EXISTS idx_hr_snapshots_company ON hr_analytics_snapshots(company_id);
CREATE INDEX IF NOT EXISTS idx_hr_snapshots_month ON hr_analytics_snapshots(company_id, snapshot_month);
