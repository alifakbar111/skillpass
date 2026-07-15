package models

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"time"
)

type Evaluation struct {
	bun.BaseModel `bun:"table:ai_evaluations"`
	ID            uuid.UUID `bun:",pk,type:uuid,default:gen_random_uuid()"`
	ProfileID     uuid.UUID `bun:",notnull"`
	OverallScore  int32     `bun:",notnull"`
	Strengths     string    `bun:",notnull"`
	Weaknesses    string    `bun:",notnull"`
	Suggestions   string    `bun:",notnull"`
	SkillScores   string    `bun:",notnull"`
	RawAnalysis   string    `bun:",notnull"`
	CreatedAt     time.Time `bun:",notnull"`
	IsCurrent     bool      `bun:",notnull"`
}
