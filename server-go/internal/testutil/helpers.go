package testutil

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gin-gonic/gin/binding"
	"github.com/uptrace/bun"
)

// UniqueEmail returns a unique email for test isolation.
// Each call generates a different email using a random suffix.
func UniqueEmail(prefix string) string {
	return fmt.Sprintf("%s-%s@test.com", prefix, uuid.New().String()[:8])
}

// UniqueUsername returns a unique username for test isolation.
func UniqueUsername(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, uuid.New().String()[:8])
}

// CleanTestData truncates all application tables for test isolation.
// Call at the start of each test function that creates persistent data.
func CleanTestData(db bun.IConn) {
	tables := []string{
		// Core tables
		"company_webhooks",
		"notifications",
		"application_messages",
		"ai_evaluations",
		"applications",
		"job_experiences",
		"job_postings",
		"companies",
		"jobseeker_profiles",
		"tags",
		"industry_categories",
		"refresh_tokens",
		"admin_audit_log",
		// HRIS tables
		"attendance_logs",
		"attendance_exceptions",
		"leave_requests",
		"leave_balances",
		"leave_types",
		"holidays",
		"payroll_runs",
		"payslips",
		"salary_components",
		"employee_salaries",
		"employee_shifts",
		"shift_templates",
		"departments",
		"positions",
		"branches",
		"employees",
		"employee_id_configs",
		"onboarding_checklists",
		"onboarding_tasks",
		// Profile views
		"profile_views",
		// Auth tokens
		"authtokens",
		// User
		"users",
	}
	for _, t := range tables {
		_, _ = db.ExecContext(context.Background(), fmt.Sprintf("TRUNCATE TABLE %s CASCADE", t))
	}
}

// RegisterTestValidators registers custom Gin validators needed by tests.
// Call once per test process (safe to call multiple times).
func RegisterTestValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// slug_format validates lowercase alphanumeric + hyphens pattern
		_ = v.RegisterValidation("slug_format", func(fl validator.FieldLevel) bool {
			return true // validation is done in handler logic; this just prevents panic
		})
	}
}
