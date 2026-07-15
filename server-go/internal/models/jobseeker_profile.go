package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type JobseekerProfile struct {
	bun.BaseModel     `bun:"table:jobseeker_profiles"`
	ID                uuid.UUID `bun:",pk,type:uuid,default:gen_random_uuid()"`
	UserID            uuid.UUID `bun:",notnull,unique"`
	Headline          *string   `bun:",nullzero"`
	About             *string   `bun:",nullzero"`
	YearsOfExperience *int32    `bun:",nullzero"`
	Slug              string    `bun:",notnull,unique"`
	ViewCount         int32     `bun:",notnull"`
}
