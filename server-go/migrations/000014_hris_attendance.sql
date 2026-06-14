-- Sprint 3: Shift config, GPS clock-in, live attendance dashboard

DO $$ BEGIN
  CREATE TYPE attendance_status AS ENUM ('present', 'absent', 'late', 'half_day', 'holiday', 'leave', 'off_day');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

DO $$ BEGIN
  CREATE TYPE exception_status AS ENUM ('pending', 'approved', 'rejected');
EXCEPTION WHEN duplicate_object THEN NULL; END $$;

CREATE TABLE IF NOT EXISTS shift_templates (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id UUID NOT NULL REFERENCES companies(id),
  name TEXT NOT NULL,
  start_time TIME NOT NULL,
  end_time TIME NOT NULL,
  break_duration_minutes INT NOT NULL DEFAULT 60,
  late_tolerance_minutes INT NOT NULL DEFAULT 15,
  overtime_multiplier NUMERIC(3,2) NOT NULL DEFAULT 1.50,
  applicable_days INT[] NOT NULL DEFAULT '{1,2,3,4,5}',
  is_default BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_shift_templates_company ON shift_templates(company_id);

CREATE TABLE IF NOT EXISTS employee_shifts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  employee_id UUID NOT NULL REFERENCES employees(id),
  shift_id UUID NOT NULL REFERENCES shift_templates(id),
  effective_date DATE NOT NULL,
  end_date DATE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_employee_shifts_employee ON employee_shifts(employee_id);

CREATE TABLE IF NOT EXISTS attendance_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id UUID NOT NULL REFERENCES companies(id),
  employee_id UUID NOT NULL REFERENCES employees(id),
  date DATE NOT NULL,
  clock_in TIMESTAMPTZ,
  clock_out TIMESTAMPTZ,
  clock_in_lat NUMERIC(10,7),
  clock_in_lng NUMERIC(10,7),
  clock_out_lat NUMERIC(10,7),
  clock_out_lng NUMERIC(10,7),
  branch_id UUID REFERENCES branches(id),
  is_in_geofence BOOLEAN,
  is_late BOOLEAN NOT NULL DEFAULT false,
  late_minutes INT NOT NULL DEFAULT 0,
  is_early_out BOOLEAN NOT NULL DEFAULT false,
  overtime_minutes INT NOT NULL DEFAULT 0,
  attendance_code TEXT NOT NULL DEFAULT 'H',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(employee_id, date)
);
CREATE INDEX IF NOT EXISTS idx_attendance_company_date ON attendance_logs(company_id, date);
CREATE INDEX IF NOT EXISTS idx_attendance_employee ON attendance_logs(employee_id);

CREATE TABLE IF NOT EXISTS attendance_exceptions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  company_id UUID NOT NULL REFERENCES companies(id),
  employee_id UUID NOT NULL REFERENCES employees(id),
  date DATE NOT NULL,
  exception_type TEXT NOT NULL,
  reason TEXT NOT NULL,
  attachment_url TEXT,
  status exception_status NOT NULL DEFAULT 'pending',
  reviewer_id UUID REFERENCES employees(id),
  reviewer_comment TEXT,
  reviewed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_exceptions_company ON attendance_exceptions(company_id);
CREATE INDEX IF NOT EXISTS idx_exceptions_employee ON attendance_exceptions(employee_id);
