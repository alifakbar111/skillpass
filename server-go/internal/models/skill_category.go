package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"time"
)

type SkillCategory struct {
	bun.BaseModel `bun:"table:skill_categories"`
	ID          uuid.UUID `bun:",pk,type:uuid,default:gen_random_uuid()"`
	Name        string    `bun:",notnull,unique"`
	Description string    `bun:",notnull"`
	CreatedAt   time.Time `bun:",notnull"`
}
