package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
)

// Service manages company webhooks and dispatches signed events.
type Service struct {
	db   *sql.DB
	http *http.Client
}

func NewService(db *sql.DB) *Service {
	return &Service{
		db: db,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type Webhook struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	Secret    string `json:"secret,omitempty"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"createdAt"`
}

// generateSecret returns a random hex secret used for HMAC signing.
func generateSecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate secret: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

func validateURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL must use http or https")
	}
	if u.Host == "" {
		return fmt.Errorf("URL must have a host")
	}
	// Resolve the hostname and reject private, loopback, and link-local addresses
	// to prevent SSRF attacks against internal network services.
	host := u.Hostname()
	ips, err := net.LookupHost(host)
	if err != nil {
		return fmt.Errorf("unable to resolve host: %q", host)
	}
	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}
		if ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return fmt.Errorf("URL must not point to a private or internal network address")
		}
	}
	return nil
}

// Create registers a webhook for a company. The generated secret is returned
// once in the response so the receiver can verify signatures.
func (s *Service) Create(ctx context.Context, companyID, rawURL string) (*Webhook, error) {
	if err := validateURL(rawURL); err != nil {
		return nil, err
	}

	secret, err := generateSecret()
	if err != nil {
		return nil, err
	}

	var id uuid.UUID
	var createdAt time.Time
	err = s.db.QueryRowContext(ctx,
		`INSERT INTO company_webhooks (company_id, url, secret)
		 VALUES ($1, $2, $3)
		 RETURNING id, created_at`,
		companyID, rawURL, secret,
	).Scan(&id, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("insert webhook: %w", err)
	}

	return &Webhook{
		ID:        id.String(),
		URL:       rawURL,
		Secret:    secret,
		Active:    true,
		CreatedAt: createdAt.Format(time.RFC3339),
	}, nil
}

// List returns a company's webhooks. Secrets are not included.
func (s *Service) List(ctx context.Context, companyID string) ([]Webhook, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, url, active, created_at
		 FROM company_webhooks
		 WHERE company_id = $1
		 ORDER BY created_at DESC`,
		companyID,
	)
	if err != nil {
		return nil, fmt.Errorf("query webhooks: %w", err)
	}
	defer rows.Close()

	webhooks := []Webhook{}
	for rows.Next() {
		var w Webhook
		var id uuid.UUID
		var createdAt time.Time
		if err := rows.Scan(&id, &w.URL, &w.Active, &createdAt); err != nil {
			return nil, fmt.Errorf("scan webhook: %w", err)
		}
		w.ID = id.String()
		w.CreatedAt = createdAt.Format(time.RFC3339)
		webhooks = append(webhooks, w)
	}
	return webhooks, rows.Err()
}

// Delete removes a webhook owned by the company. Returns false if not found.
func (s *Service) Delete(ctx context.Context, webhookID, companyID string) (bool, error) {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM company_webhooks WHERE id = $1 AND company_id = $2`,
		webhookID, companyID,
	)
	if err != nil {
		return false, fmt.Errorf("delete webhook: %w", err)
	}
	ra, _ := res.RowsAffected()
	return ra > 0, nil
}

// Event is the payload posted to webhook receivers.
type Event struct {
	Type      string          `json:"type"`
	Timestamp string          `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// DispatchApplicationReceived posts an "application.received" event to all active
// webhooks of the company that owns the job. Runs asynchronously — call it and move on.
func (s *Service) DispatchApplicationReceived(ctx context.Context, jobPostingID, jobseekerProfileID string) error {
	// Look up the owning company and event context first (synchronously, request ctx).
	var companyID uuid.UUID
	var jobTitle string
	err := s.db.QueryRowContext(ctx,
		`SELECT company_id, title FROM job_postings WHERE id = $1`, jobPostingID,
	).Scan(&companyID, &jobTitle)
	if err != nil {
		return fmt.Errorf("lookup job: %w", err)
	}

	var candidateName string
	err = s.db.QueryRowContext(ctx,
		`SELECT u.name FROM jobseeker_profiles jp JOIN users u ON u.id = jp.user_id WHERE jp.id = $1`,
		jobseekerProfileID,
	).Scan(&candidateName)
	if err != nil {
		return fmt.Errorf("lookup candidate: %w", err)
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT url, secret FROM company_webhooks WHERE company_id = $1 AND active = TRUE`,
		companyID,
	)
	if err != nil {
		return fmt.Errorf("query webhooks: %w", err)
	}
	defer rows.Close()

	type target struct{ url, secret string }
	var targets []target
	for rows.Next() {
		var t target
		if err := rows.Scan(&t.url, &t.secret); err != nil {
			return fmt.Errorf("scan webhook: %w", err)
		}
		targets = append(targets, t)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(targets) == 0 {
		return nil
	}

	data, _ := json.Marshal(map[string]string{
		"jobPostingId":  jobPostingID,
		"jobTitle":      jobTitle,
		"candidateName": candidateName,
	})
	payload, _ := json.Marshal(Event{
		Type:      "application.received",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data:      data,
	})

	// Fire-and-forget: deliver outside the request lifecycle.
	for _, t := range targets {
		go s.deliver(t.url, t.secret, payload)
	}
	return nil
}

// deliver posts the payload with an HMAC-SHA256 signature header.
// Receivers verify with: hex(hmac_sha256(secret, body)).
func (s *Service) deliver(targetURL, secret string, payload []byte) {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(payload))
	if err != nil {
		slog.Warn("webhook request build failed", "url", targetURL, "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-SkillPass-Signature", signature)

	resp, err := s.http.Do(req)
	if err != nil {
		slog.Warn("webhook delivery failed", "url", targetURL, "error", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		slog.Warn("webhook receiver returned non-2xx", "url", targetURL, "status", resp.StatusCode)
	}
}
