package ats

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ---------- Offer templates ----------

type OfferTemplateResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Body      string `json:"body"`
	CreatedAt string `json:"createdAt"`
} //@name AtsOfferTemplate

type OfferTemplateInput struct {
	Name string `json:"name"`
	Body string `json:"body"`
}

func (s *Service) ListOfferTemplates(ctx context.Context, companyID uuid.UUID) ([]OfferTemplateResponse, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, body, created_at FROM ats_offer_templates WHERE company_id=$1 ORDER BY created_at DESC`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []OfferTemplateResponse{}
	for rows.Next() {
		var t OfferTemplateResponse
		var created time.Time
		if err := rows.Scan(&t.ID, &t.Name, &t.Body, &created); err != nil {
			return nil, err
		}
		t.CreatedAt = created.Format(time.RFC3339)
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Service) CreateOfferTemplate(ctx context.Context, companyID uuid.UUID, in OfferTemplateInput) (*OfferTemplateResponse, error) {
	if in.Name == "" || in.Body == "" {
		return nil, ErrBadInput
	}
	var t OfferTemplateResponse
	var created time.Time
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO ats_offer_templates (company_id, name, body) VALUES ($1,$2,$3) RETURNING id, name, body, created_at`,
		companyID, in.Name, in.Body).Scan(&t.ID, &t.Name, &t.Body, &created)
	if err != nil {
		return nil, err
	}
	t.CreatedAt = created.Format(time.RFC3339)
	return &t, nil
}

func (s *Service) UpdateOfferTemplate(ctx context.Context, companyID, templateID uuid.UUID, in OfferTemplateInput) (*OfferTemplateResponse, error) {
	if in.Name == "" || in.Body == "" {
		return nil, ErrBadInput
	}
	var t OfferTemplateResponse
	var created time.Time
	err := s.db.QueryRowContext(ctx,
		`UPDATE ats_offer_templates SET name=$3, body=$4 WHERE id=$1 AND company_id=$2 RETURNING id, name, body, created_at`,
		templateID, companyID, in.Name, in.Body).Scan(&t.ID, &t.Name, &t.Body, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	t.CreatedAt = created.Format(time.RFC3339)
	return &t, nil
}

func (s *Service) DeleteOfferTemplate(ctx context.Context, companyID, templateID uuid.UUID) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM ats_offer_templates WHERE id=$1 AND company_id=$2`, templateID, companyID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// ---------- Offer letters ----------

type OfferResponse struct {
	ID            string  `json:"id"`
	CandidateID   string  `json:"candidateId"`
	TemplateID    *string `json:"templateId,omitempty"`
	PositionTitle string  `json:"positionTitle"`
	Salary        *string `json:"salary,omitempty"`
	StartDate     *string `json:"startDate,omitempty"`
	Body          string  `json:"body"`
	Status        string  `json:"status"`
	AcceptToken   string  `json:"acceptToken,omitempty"` // company-side only; blank on public view
	SignatureName *string `json:"signatureName,omitempty"`
	SignedAt      *string `json:"signedAt,omitempty"`
	SentAt        *string `json:"sentAt,omitempty"`
	CreatedAt     string  `json:"createdAt"`
} //@name AtsOffer

type OfferInput struct {
	TemplateID    *string `json:"templateId,omitempty"`
	PositionTitle string  `json:"positionTitle"`
	Salary        *string `json:"salary,omitempty"`
	StartDate     *string `json:"startDate,omitempty"`
	Body          string  `json:"body"` // optional; if empty and templateId set, render from template
}

// CreateOffer drafts an offer for a candidate. If a template is provided and no
// body is given, merge fields are substituted from the offer + candidate data.
func (s *Service) CreateOffer(ctx context.Context, companyID, candidateID, createdBy uuid.UUID, in OfferInput) (*OfferResponse, error) {
	if in.PositionTitle == "" {
		return nil, ErrBadInput
	}
	// Load candidate (also enforces company scope).
	cand, err := s.GetCandidate(ctx, companyID, candidateID)
	if err != nil {
		return nil, err
	}

	body := in.Body
	if body == "" && in.TemplateID != nil && *in.TemplateID != "" {
		tid, perr := uuid.Parse(*in.TemplateID)
		if perr != nil {
			return nil, ErrBadInput
		}
		var tplBody string
		if err := s.db.QueryRowContext(ctx,
			`SELECT body FROM ats_offer_templates WHERE id=$1 AND company_id=$2`, tid, companyID).Scan(&tplBody); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrNotFound
			}
			return nil, err
		}
		body = renderTemplate(tplBody, cand, in)
	}
	if body == "" {
		return nil, ErrBadInput
	}

	token, err := newToken()
	if err != nil {
		return nil, err
	}

	var id uuid.UUID
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO ats_offer_letters (candidate_id, template_id, position_title, salary, start_date, body, accept_token, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		candidateID, in.TemplateID, in.PositionTitle, nullDecimal(in.Salary), nullDate(in.StartDate), body, token, createdBy).Scan(&id)
	if err != nil {
		return nil, err
	}
	return s.getOffer(ctx, companyID, id, true)
}

// SendOffer marks a draft offer as sent and notifies the candidate's user.
func (s *Service) SendOffer(ctx context.Context, companyID, offerID uuid.UUID) (*OfferResponse, error) {
	res, err := s.db.ExecContext(ctx, `
		UPDATE ats_offer_letters o SET status='sent', sent_at=now()
		WHERE o.id=$1 AND o.status='draft'
		  AND EXISTS (SELECT 1 FROM ats_candidates c WHERE c.id=o.candidate_id AND c.company_id=$2)`,
		offerID, companyID)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, ErrNotFound
	}
	var candidateID uuid.UUID
	_ = s.db.QueryRowContext(ctx, `SELECT candidate_id FROM ats_offer_letters WHERE id=$1`, offerID).Scan(&candidateID)
	s.notifyCandidateUser(ctx, candidateID, "offer", "You received an offer",
		"A job offer is waiting for your review and signature.")
	return s.getOffer(ctx, companyID, offerID, true)
}

func (s *Service) ListOffers(ctx context.Context, companyID, candidateID uuid.UUID) ([]OfferResponse, error) {
	if err := s.assertCandidate(ctx, companyID, candidateID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, candidate_id, template_id, position_title, salary::text, start_date::text, body, status::text,
		       accept_token, signature_name, signed_at, sent_at, created_at
		FROM ats_offer_letters WHERE candidate_id=$1 ORDER BY created_at DESC`, candidateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []OfferResponse{}
	for rows.Next() {
		o, err := scanOffer(rows, true)
		if err != nil {
			return nil, err
		}
		out = append(out, *o)
	}
	return out, rows.Err()
}

func (s *Service) getOffer(ctx context.Context, companyID, offerID uuid.UUID, withToken bool) (*OfferResponse, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT o.id, o.candidate_id, o.template_id, o.position_title, o.salary::text, o.start_date::text, o.body, o.status::text,
		       o.accept_token, o.signature_name, o.signed_at, o.sent_at, o.created_at
		FROM ats_offer_letters o
		JOIN ats_candidates c ON c.id = o.candidate_id
		WHERE o.id=$1 AND c.company_id=$2`, offerID, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, ErrNotFound
	}
	return scanOffer(rows, withToken)
}

func scanOffer(rows *sql.Rows, withToken bool) (*OfferResponse, error) {
	var o OfferResponse
	var token string
	var salary, startDate *string
	var signedAt, sentAt *time.Time
	var created time.Time
	if err := rows.Scan(&o.ID, &o.CandidateID, &o.TemplateID, &o.PositionTitle, &salary, &startDate,
		&o.Body, &o.Status, &token, &o.SignatureName, &signedAt, &sentAt, &created); err != nil {
		return nil, err
	}
	o.Salary = salary
	o.StartDate = startDate
	if withToken {
		o.AcceptToken = token
	}
	if signedAt != nil {
		v := signedAt.Format(time.RFC3339)
		o.SignedAt = &v
	}
	if sentAt != nil {
		v := sentAt.Format(time.RFC3339)
		o.SentAt = &v
	}
	o.CreatedAt = created.Format(time.RFC3339)
	return &o, nil
}

// ---------- Public accept (candidate token) ----------

// PublicOfferResponse is the candidate-facing view of an offer (no token, no
// internal ids).
type PublicOfferResponse struct {
	CompanyName   string  `json:"companyName"`
	CandidateName string  `json:"candidateName"`
	PositionTitle string  `json:"positionTitle"`
	Salary        *string `json:"salary,omitempty"`
	StartDate     *string `json:"startDate,omitempty"`
	Body          string  `json:"body"`
	Status        string  `json:"status"`
} //@name PublicOffer

// GetPublicOffer resolves an offer by its accept token for the public page.
func (s *Service) GetPublicOffer(ctx context.Context, token string) (*PublicOfferResponse, error) {
	if token == "" {
		return nil, ErrNotFound
	}
	var p PublicOfferResponse
	var salary, startDate *string
	err := s.db.QueryRowContext(ctx, `
		SELECT co.company_name, c.candidate_name, o.position_title, o.salary::text, o.start_date::text, o.body, o.status::text
		FROM ats_offer_letters o
		JOIN ats_candidates c ON c.id = o.candidate_id
		JOIN companies co ON co.id = c.company_id
		WHERE o.accept_token=$1`, token).
		Scan(&p.CompanyName, &p.CandidateName, &p.PositionTitle, &salary, &startDate, &p.Body, &p.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p.Salary = salary
	p.StartDate = startDate
	return &p, nil
}

// AcceptOfferResult reports the outcome of a candidate acceptance.
type AcceptOfferResult struct {
	Status         string `json:"status"`
	EmployeeLinked bool   `json:"employeeLinked"` // true if bound to the candidate's existing login
	Message        string `json:"message"`
} //@name AcceptOfferResult

// AcceptOffer is the public, token-authenticated acceptance. It records the
// typed signature, marks the offer accepted, sets the candidate to hired, and
// runs the ATS→HRIS bridge (create/link the employee record).
func (s *Service) AcceptOffer(ctx context.Context, token, signatureName string) (*AcceptOfferResult, error) {
	signatureName = strings.TrimSpace(signatureName)
	if token == "" || signatureName == "" {
		return nil, ErrBadInput
	}

	var offerID, candidateID uuid.UUID
	var status string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, candidate_id, status::text FROM ats_offer_letters WHERE accept_token=$1`, token).
		Scan(&offerID, &candidateID, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if status == "accepted" {
		return &AcceptOfferResult{Status: "accepted", Message: "This offer has already been accepted."}, nil
	}
	if status == "declined" || status == "expired" {
		return nil, ErrBadInput
	}

	if _, err := s.db.ExecContext(ctx,
		`UPDATE ats_offer_letters SET status='accepted', signature_name=$2, signed_at=now() WHERE id=$1`,
		offerID, signatureName); err != nil {
		return nil, err
	}
	if _, err := s.db.ExecContext(ctx,
		`UPDATE ats_candidates SET status='hired', updated_at=now() WHERE id=$1`, candidateID); err != nil {
		return nil, err
	}

	// Bridge into HRIS (best-effort but reported).
	linked, bridgeErr := s.onboardHiredCandidate(ctx, candidateID)
	msg := "Offer accepted. Welcome aboard!"
	if bridgeErr != nil {
		msg = "Offer accepted. Your employee profile will be finalized by HR."
	}
	return &AcceptOfferResult{Status: "accepted", EmployeeLinked: linked, Message: msg}, nil
}

// DeclineOffer is the public decline path.
func (s *Service) DeclineOffer(ctx context.Context, token string) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE ats_offer_letters SET status='declined' WHERE accept_token=$1 AND status IN ('sent','draft')`, token)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// ---------- helpers ----------

func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func renderTemplate(body string, cand *CandidateResponse, in OfferInput) string {
	repl := map[string]string{
		"{{candidateName}}": cand.CandidateName,
		"{{position}}":      in.PositionTitle,
		"{{salary}}":        derefStr(in.Salary),
		"{{startDate}}":     derefStr(in.StartDate),
	}
	for k, v := range repl {
		body = strings.ReplaceAll(body, k, v)
	}
	return body
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func nullDecimal(p *string) any {
	if p == nil || *p == "" {
		return nil
	}
	return *p
}

func nullDate(p *string) any {
	if p == nil || *p == "" {
		return nil
	}
	return *p
}
