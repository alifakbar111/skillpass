package models

import (
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/uptrace/bun"
)

type JobExperience struct {
	bun.BaseModel `bun:"table:job_experiences"`
	ID            uuid.UUID      `bun:",pk,type:uuid,default:gen_random_uuid()"`
	ProfileID     uuid.UUID      `bun:",notnull"`
	Type          string         `bun:",notnull"`
	Title         string         `bun:",notnull"`
	Organization  string         `bun:",notnull"`
	StartDate     string         `bun:",notnull"`
	EndDate       *string        `bun:",nullzero"`
	IsCurrent     bool           `bun:",notnull"`
	Description   *string        `bun:",nullzero"`
	Industry      *string        `bun:",nullzero"`
	SkillsUsed    *pq.StringArray `bun:",nullzero"`
	URL           *string        `bun:",nullzero"`
	SortOrder     int32          `bun:",notnull"`
}
