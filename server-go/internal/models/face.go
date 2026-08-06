package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type FaceEnrollment struct {
	bun.BaseModel   `bun:"table:face_enrollments"`
	ID              uuid.UUID  `bun:",pk,type:uuid,default:gen_random_uuid()"`
	EmployeeID      uuid.UUID  `bun:",notnull"`
	EmbeddingVector []byte     `bun:",notnull"` // encrypted (AES-256-GCM)
	LivenessScore   *float64   `bun:",nullzero"`
	IsActive        bool       `bun:",notnull"`
	EnrolledBy      *uuid.UUID `bun:",nullzero"`
	EnrolledAt      time.Time  `bun:",notnull"`
}

type FaceVerificationLog struct {
	bun.BaseModel `bun:"table:face_verification_logs"`
	ID            uuid.UUID `bun:",pk,type:uuid,default:gen_random_uuid()"`
	EmployeeID    uuid.UUID `bun:",notnull"`
	Action        string    `bun:",notnull"`
	MatchScore    *float64  `bun:",nullzero"`
	LivenessScore *float64  `bun:",nullzero"`
	Passed        bool      `bun:",notnull"`
	IPAddress     *string   `bun:",nullzero"`
	UserAgent     *string   `bun:",nullzero"`
	CreatedAt     time.Time `bun:",notnull"`
}
