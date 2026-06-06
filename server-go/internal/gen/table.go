package gen

import (
	"skillpass-server-go/.gen/skillpass/public/table"
)

// Table references for type-safe SQL building.
var (
	Users              = table.Users
	Companies          = table.Companies
	JobseekerProfiles  = table.JobseekerProfiles
	JobExperiences     = table.JobExperiences
	IndustryCategories = table.IndustryCategories
	Tags               = table.Tags
	JobPostings        = table.JobPostings
	RefreshTokens      = table.RefreshTokens
	AdminAuditLog      = table.AdminAuditLog
)
