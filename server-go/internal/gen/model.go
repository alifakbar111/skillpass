package gen

import (
	"skillpass-server-go/.gen/skillpass/public/model"
)

//go:generate jet -dsn=postgres://postgres:postgres@localhost:5432/skillpass?sslmode=disable -schema=public -path=../../.gen -ignore-tables=__drizzle_migrations

// Model type aliases for query result mapping.
type (
	User              = model.Users
	Company           = model.Companies
	JobseekerProfile  = model.JobseekerProfiles
	JobExperience     = model.JobExperiences
	IndustryCategory  = model.IndustryCategories
	Tag               = model.Tags
	JobPosting        = model.JobPostings
	RefreshToken      = model.RefreshTokens
	AdminAudit        = model.AdminAuditLog
)
