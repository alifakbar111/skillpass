package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type JobCategoryWeight struct {
	bun.BaseModel `bun:"table:job_category_weights"`
	ID            uuid.UUID `bun:",pk,type:uuid,default:gen_random_uuid()"`
	JobPostingID  uuid.UUID `bun:",notnull"`
	CategoryID    uuid.UUID `bun:",notnull"`
	Weight        int32     `bun:",notnull"`
}
