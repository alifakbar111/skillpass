-- server-go/migrations/000012_hris_foundation.sql
-- HRIS Foundation: RBAC, Org Structure, Employee DB

-- ============================================================
-- ENUMS
-- ============================================================

DO $$ BEGIN
  CREATE TYPE employment_type AS ENUM ('permanent', 'contract', 'probation', 'intern');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
  CREATE TYPE employment_status AS ENUM ('active', 'resigned', 'terminated', 'on_leave');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
  CREATE TYPE job_level AS ENUM ('staff', 'supervisor', 'manager', 'director');
EXCEPTION WHEN duplicate_object THEN null;
END $$;

-- ============================================================
-- ORG STRUCTURE
-- ============================================================

CREATE TABLE IF NOT EXISTS branches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    branch_type TEXT NOT NULL DEFAULT 'branch',
    parent_branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
    address TEXT,
    city TEXT,
    province TEXT,
    latitude NUMERIC(10,7),
    longitude NUMERIC(10,7),
    geofence_radius_meters INTEGER NOT NULL DEFAULT 200,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS branches_company_idx ON branches(company_id);

CREATE TABLE IF NOT EXISTS departments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    parent_department_id UUID REFERENCES departments(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS departments_company_idx ON departments(company_id);

CREATE TABLE IF NOT EXISTS positions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    department_id UUID REFERENCES departments(id) ON DELETE SET NULL,
    level job_level NOT NULL DEFAULT 'staff',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS positions_company_idx ON positions(company_id);
CREATE INDEX IF NOT EXISTS positions_department_idx ON positions(department_id);

-- ============================================================
-- EMPLOYEES
-- ============================================================

CREATE TABLE IF NOT EXISTS employees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    employee_id_number TEXT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL,
    phone TEXT,
    date_of_birth DATE,
    gender TEXT,
    marital_status TEXT,
    address TEXT,
    city TEXT,
    province TEXT,
    postal_code TEXT,
    national_id TEXT,
    npwp TEXT,
    bpjs_kesehatan_id TEXT,
    bpjs_ketenagakerjaan_id TEXT,
    bank_name TEXT,
    bank_account_number TEXT,
    bank_account_holder TEXT,
    emergency_contact_name TEXT,
    emergency_contact_phone TEXT,
    emergency_contact_relation TEXT,
    employment_type employment_type NOT NULL DEFAULT 'permanent',
    employment_status employment_status NOT NULL DEFAULT 'active',
    join_date DATE NOT NULL,
    end_date DATE,
    department_id UUID REFERENCES departments(id) ON DELETE SET NULL,
    position_id UUID REFERENCES positions(id) ON DELETE SET NULL,
    branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
    manager_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    base_salary NUMERIC(15,2),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(company_id, employee_id_number)
);

CREATE INDEX IF NOT EXISTS employees_company_idx ON employees(company_id);
CREATE INDEX IF NOT EXISTS employees_user_idx ON employees(user_id);
CREATE INDEX IF NOT EXISTS employees_department_idx ON employees(department_id);
CREATE INDEX IF NOT EXISTS employees_branch_idx ON employees(branch_id);
CREATE INDEX IF NOT EXISTS employees_manager_idx ON employees(manager_id);
CREATE INDEX IF NOT EXISTS employees_status_idx ON employees(company_id, employment_status);

CREATE TABLE IF NOT EXISTS employee_id_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL UNIQUE REFERENCES companies(id) ON DELETE CASCADE,
    prefix TEXT NOT NULL DEFAULT 'EMP',
    next_sequence INTEGER NOT NULL DEFAULT 1,
    padding INTEGER NOT NULL DEFAULT 4
);

-- ============================================================
-- RBAC
-- ============================================================

CREATE TABLE IF NOT EXISTS hris_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(company_id, name)
);

CREATE INDEX IF NOT EXISTS hris_roles_company_idx ON hris_roles(company_id);

CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code TEXT NOT NULL UNIQUE,
    module TEXT NOT NULL,
    description TEXT
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID NOT NULL REFERENCES hris_roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS employee_roles (
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES hris_roles(id) ON DELETE CASCADE,
    PRIMARY KEY (employee_id, role_id)
);

-- ============================================================
-- SEED: Permissions (global, not per-company)
-- ============================================================

INSERT INTO permissions (id, code, module, description) VALUES
    -- Employee module
    (gen_random_uuid(), 'employee.view',       'employee',    'View employee records'),
    (gen_random_uuid(), 'employee.view_team',   'employee',    'View own team employee records'),
    (gen_random_uuid(), 'employee.view_self',   'employee',    'View own employee record'),
    (gen_random_uuid(), 'employee.create',      'employee',    'Create new employees'),
    (gen_random_uuid(), 'employee.update',      'employee',    'Update employee records'),
    (gen_random_uuid(), 'employee.delete',      'employee',    'Delete/terminate employees'),
    -- Attendance module
    (gen_random_uuid(), 'attendance.view',       'attendance',  'View all attendance records'),
    (gen_random_uuid(), 'attendance.view_team',  'attendance',  'View team attendance records'),
    (gen_random_uuid(), 'attendance.view_self',  'attendance',  'View own attendance'),
    (gen_random_uuid(), 'attendance.manage',     'attendance',  'Manage shifts and attendance settings'),
    (gen_random_uuid(), 'attendance.clock',      'attendance',  'Clock in/out'),
    (gen_random_uuid(), 'attendance.approve',    'attendance',  'Approve attendance exceptions'),
    (gen_random_uuid(), 'attendance.export',     'attendance',  'Export attendance reports'),
    -- Leave module
    (gen_random_uuid(), 'leave.view',            'leave',       'View all leave records'),
    (gen_random_uuid(), 'leave.view_team',       'leave',       'View team leave records'),
    (gen_random_uuid(), 'leave.request',         'leave',       'Submit leave requests'),
    (gen_random_uuid(), 'leave.approve',         'leave',       'Approve leave requests'),
    (gen_random_uuid(), 'leave.manage',          'leave',       'Manage leave types and policies'),
    -- Payroll module
    (gen_random_uuid(), 'payroll.view',          'payroll',     'View all payroll data'),
    (gen_random_uuid(), 'payroll.view_self',     'payroll',     'View own payslip'),
    (gen_random_uuid(), 'payroll.run',           'payroll',     'Run payroll calculations'),
    (gen_random_uuid(), 'payroll.approve',       'payroll',     'Approve payroll runs'),
    (gen_random_uuid(), 'payroll.manage',        'payroll',     'Manage payroll components and config'),
    -- Performance/KPI module
    (gen_random_uuid(), 'performance.view',      'performance', 'View all performance data'),
    (gen_random_uuid(), 'performance.view_team', 'performance', 'View team performance'),
    (gen_random_uuid(), 'performance.view_self', 'performance', 'View own KPI scores'),
    (gen_random_uuid(), 'performance.manage',    'performance', 'Manage KPI frameworks and reviews'),
    (gen_random_uuid(), 'performance.review',    'performance', 'Review and score team members'),
    -- ATS module
    (gen_random_uuid(), 'ats.view',              'ats',         'View ATS pipeline'),
    (gen_random_uuid(), 'ats.manage',            'ats',         'Manage job postings and pipeline'),
    (gen_random_uuid(), 'ats.scorecard',         'ats',         'Submit interview scorecards'),
    -- Analytics module
    (gen_random_uuid(), 'analytics.view',        'analytics',   'View full analytics'),
    (gen_random_uuid(), 'analytics.view_team',   'analytics',   'View team-level analytics'),
    (gen_random_uuid(), 'analytics.view_exec',   'analytics',   'View executive dashboard'),
    (gen_random_uuid(), 'analytics.export',      'analytics',   'Export analytics reports'),
    -- Documents module
    (gen_random_uuid(), 'documents.view',        'documents',   'View all company documents'),
    (gen_random_uuid(), 'documents.view_team',   'documents',   'View team member documents'),
    (gen_random_uuid(), 'documents.view_self',   'documents',   'View own documents'),
    (gen_random_uuid(), 'documents.manage',      'documents',   'Upload and manage documents'),
    -- Org/Settings module
    (gen_random_uuid(), 'org.view',              'org',         'View org chart and structure'),
    (gen_random_uuid(), 'org.manage',            'org',         'Manage branches, departments, positions'),
    (gen_random_uuid(), 'roles.manage',          'org',         'Manage HRIS roles and permissions'),
    (gen_random_uuid(), 'settings.manage',       'org',         'Manage company HRIS settings')
ON CONFLICT (code) DO NOTHING;
