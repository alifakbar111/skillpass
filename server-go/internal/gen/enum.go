package gen

import (
	"skillpass-server-go/.gen/skillpass/public/enum"
)

// Enum value accessors.
var (
	RoleJobseeker = enum.Role.Jobseeker
	RoleCompany   = enum.Role.Company
	RoleAdmin     = enum.Role.Admin

	VerificationStatusPending  = enum.VerificationStatus.Pending
	VerificationStatusVerified = enum.VerificationStatus.Verified
	VerificationStatusRejected = enum.VerificationStatus.Rejected

	JobStatusOpen   = enum.JobStatus.Open
	JobStatusClosed = enum.JobStatus.Closed

	ExperienceLevelEntry  = enum.ExperienceLevel.Entry
	ExperienceLevelMid    = enum.ExperienceLevel.Mid
	ExperienceLevelSenior = enum.ExperienceLevel.Senior
	ExperienceLevelLead   = enum.ExperienceLevel.Lead

	ExperienceTypeEmployment    = enum.ExperienceType.Employment
	ExperienceTypeGig           = enum.ExperienceType.Gig
	ExperienceTypeEducation     = enum.ExperienceType.Education
	ExperienceTypeCertification = enum.ExperienceType.Certification
	ExperienceTypeProject       = enum.ExperienceType.Project
	ExperienceTypeVolunteering  = enum.ExperienceType.Volunteering
)
