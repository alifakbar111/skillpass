package face

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	"skillpass-server-go/internal/models"
)

var (
	ErrDisabled         = errors.New("face recognition is not configured")
	ErrEmployeeNotFound = errors.New("employee not found")
	ErrNotEnrolled      = errors.New("no active face enrollment")
)

type Service struct {
	db           *sql.DB
	bun          bun.IDB
	client       *Client
	crypto       *cryptor
	matchThresh  float64
	reviewThresh float64
}

func NewService(db *sql.DB, bunDB bun.IDB, client *Client, secret string, matchThresh, reviewThresh float64) (*Service, error) {
	cr, err := newCryptor(secret)
	if err != nil {
		return nil, err
	}
	return &Service{db: db, bun: bunDB, client: client, crypto: cr, matchThresh: matchThresh, reviewThresh: reviewThresh}, nil
}

// requireBiometricConsent is a Sprint-8 hook. UU PDP requires explicit consent
// before collecting biometric data; this is a no-op until internal/privacy
// (Sprint 8) wires real consent records.
func (s *Service) requireBiometricConsent(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (s *Service) employeeInCompany(ctx context.Context, companyID, employeeID uuid.UUID) (bool, error) {
	var ok bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM employees WHERE id = $1 AND company_id = $2)`, employeeID, companyID).Scan(&ok)
	return ok, err
}

type EnrollResponse struct {
	Enrolled      bool    `json:"enrolled"`
	LivenessScore float64 `json:"livenessScore"`
	EnrolledAt    string  `json:"enrolledAt"`
} //@name FaceEnrollResponse

// Enroll captures a face embedding for an employee, encrypts it, and replaces
// any previous active enrollment.
func (s *Service) Enroll(ctx context.Context, companyID, employeeID uuid.UUID, enrolledBy *uuid.UUID, image []byte) (*EnrollResponse, error) {
	if !s.client.Enabled() {
		return nil, ErrDisabled
	}
	if err := s.requireBiometricConsent(ctx, employeeID); err != nil {
		return nil, err
	}
	if ok, err := s.employeeInCompany(ctx, companyID, employeeID); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrEmployeeNotFound
	}

	res, err := s.client.Enroll(ctx, image)
	if err != nil {
		return nil, err
	}
	enc, err := s.crypto.encrypt(res.Embedding)
	if err != nil {
		return nil, fmt.Errorf("encrypt embedding: %w", err)
	}

	now := time.Now()
	if _, err := s.db.ExecContext(ctx,
		`UPDATE face_enrollments SET is_active = false WHERE employee_id = $1 AND is_active = true`, employeeID); err != nil {
		return nil, fmt.Errorf("deactivate old enrollment: %w", err)
	}
	enrollment := &models.FaceEnrollment{
		EmployeeID:      employeeID,
		EmbeddingVector: enc,
		LivenessScore:   &res.Liveness,
		IsActive:        true,
		EnrolledBy:      enrolledBy,
		EnrolledAt:      now,
	}
	if _, err := s.bun.NewInsert().Model(enrollment).Exec(ctx); err != nil {
		return nil, fmt.Errorf("insert enrollment: %w", err)
	}
	s.log(ctx, employeeID, "enroll", nil, &res.Liveness, true, "", "")

	return &EnrollResponse{Enrolled: true, LivenessScore: res.Liveness, EnrolledAt: now.Format(time.RFC3339)}, nil
}

type StatusResponse struct {
	Enrolled      bool     `json:"enrolled"`
	EnrolledAt    string   `json:"enrolledAt,omitempty"`
	LivenessScore *float64 `json:"livenessScore,omitempty"`
} //@name FaceStatusResponse

// Status reports whether an employee has an active enrollment.
func (s *Service) Status(ctx context.Context, companyID, employeeID uuid.UUID) (*StatusResponse, error) {
	if ok, err := s.employeeInCompany(ctx, companyID, employeeID); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrEmployeeNotFound
	}

	var enrolledAt time.Time
	var liveness *float64
	err := s.db.QueryRowContext(ctx,
		`SELECT enrolled_at, liveness_score FROM face_enrollments WHERE employee_id = $1 AND is_active = true`,
		employeeID).Scan(&enrolledAt, &liveness)
	if errors.Is(err, sql.ErrNoRows) {
		return &StatusResponse{Enrolled: false}, nil
	}
	if err != nil {
		return nil, err
	}
	return &StatusResponse{Enrolled: true, EnrolledAt: enrolledAt.Format(time.RFC3339), LivenessScore: liveness}, nil
}

// Decision is the outcome of a verification against the configured thresholds.
type Decision struct {
	Match    float64 `json:"matchScore"`
	Liveness float64 `json:"livenessScore"`
	Outcome  string  `json:"outcome"` // accept | review | reject
}

// Verify checks a live image against the stored embedding. Used by attendance
// clock-in (Sprint 3) and proctoring (Sprint 7).
func (s *Service) Verify(ctx context.Context, employeeID uuid.UUID, image []byte, action, ip, ua string) (*Decision, error) {
	if !s.client.Enabled() {
		return nil, ErrDisabled
	}
	var enc []byte
	err := s.db.QueryRowContext(ctx,
		`SELECT embedding_vector FROM face_enrollments WHERE employee_id = $1 AND is_active = true`, employeeID).Scan(&enc)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotEnrolled
	}
	if err != nil {
		return nil, err
	}
	embedding, err := s.crypto.decrypt(enc)
	if err != nil {
		return nil, fmt.Errorf("decrypt embedding: %w", err)
	}

	res, err := s.client.Verify(ctx, image, embedding)
	if err != nil {
		return nil, err
	}

	outcome := "reject"
	switch {
	case res.Liveness < s.reviewThresh:
		outcome = "reject" // failed liveness
	case res.Match >= s.matchThresh:
		outcome = "accept"
	case res.Match >= s.reviewThresh:
		outcome = "review"
	}
	passed := outcome == "accept"
	s.log(ctx, employeeID, action, &res.Match, &res.Liveness, passed, ip, ua)

	return &Decision{Match: res.Match, Liveness: res.Liveness, Outcome: outcome}, nil
}

func (s *Service) log(ctx context.Context, employeeID uuid.UUID, action string, match, liveness *float64, passed bool, ip, ua string) {
	var ipPtr, uaPtr *string
	if ip != "" {
		ipPtr = &ip
	}
	if ua != "" {
		uaPtr = &ua
	}
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO face_verification_logs (employee_id, action, match_score, liveness_score, passed, ip_address, user_agent)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		employeeID, action, match, liveness, passed, ipPtr, uaPtr); err != nil {
		// best-effort audit; never fail the caller on log errors
		_ = err
	}
}
