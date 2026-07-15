package models

import (
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/uptrace/bun"
	"time"
)

type JobPosting struct {
	bun.BaseModel       `bun:"table:job_postings"`
	ID                  uuid.UUID      `bun:",pk,type:uuid,default:gen_random_uuid()"`
	CompanyID           uuid.UUID      `bun:",notnull"`
	Title               string         `bun:",notnull"`
	Description         string         `bun:",notnull"`
	Industry            string         `bun:",notnull"`
	Tags                *pq.StringArray `bun:",nullzero"`
	RequiredSkills      *pq.StringArray `bun:",nullzero"`
	ExperienceLevel     *string        `bun:",nullzero"`
	Location            *string        `bun:",nullzero"`
	SalaryRange         *string        `bun:",nullzero"`
	Status              string         `bun:",notnull"`
	CreatedAt           time.Time      `bun:",notnull"`
	UpdatedAt           time.Time      `bun:",notnull"`
	Requirements        *string        `bun:",nullzero"`
	Benefits            *string        `bun:",nullzero"`
	YearsExperienceMin  *int32         `bun:",nullzero"`
	YearsExperienceMax  *int32         `bun:",nullzero"`
	IsFreshGradFriendly bool           `bun:",notnull"`
}
