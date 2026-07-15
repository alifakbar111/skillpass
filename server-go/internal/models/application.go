package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"time"
)

type Application struct {
	bun.BaseModel `bun:"table:applications"`
	ID            uuid.UUID `bun:",pk,type:uuid,default:gen_random_uuid()"`
	JobseekerID   uuid.UUID `bun:",notnull"`
	JobPostingID  uuid.UUID `bun:",notnull"`
	Status        string    `bun:",notnull"`
	CreatedAt     time.Time `bun:",notnull"`
	UpdatedAt     time.Time `bun:",notnull"`
}
