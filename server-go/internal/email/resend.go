package email

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/resend/resend-go/v3"
)

// Compile-time interface check.
var _ Sender = (*ResendSender)(nil)

// ResendSender sends email via the Resend HTTP API.
// Uses RESEND_API_KEY environment variable for authentication.
// Safe for concurrent use.
type ResendSender struct {
	client *resend.Client
	from   string
}

// NewResendSender creates a ResendSender using the given API key.
// The from address is read from EMAIL_FROM (default: no-reply@skillpass.local).
func NewResendSender(apiKey string) *ResendSender {
	from := os.Getenv("EMAIL_FROM")
	if from == "" {
		from = "no-reply@skillpass.local"
	}
	client := resend.NewClient(apiKey)
	return &ResendSender{client: client, from: from}
}

// Send delivers an email via the Resend HTTP API.
// Logs the response ID on success or the error on failure.
func (s *ResendSender) Send(ctx context.Context, to, subject, htmlBody, textBody string) error {
	params := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
		Text:    textBody,
	}

	resp, err := s.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		slog.Error("resend email delivery failed",
			"to", to,
			"subject", subject,
			"error", err,
		)
		return fmt.Errorf("resend send to %s: %w", to, err)
	}

	slog.Debug("resend email sent",
		"to", to,
		"subject", subject,
		"id", resp.Id,
	)
	return nil
}

// ResendAPIKey returns the Resend API key from the environment.
// It first checks RESEND_API_KEY, then falls back to SMTP_PASS
// if it starts with "re_" (Resend API key prefix).
func ResendAPIKey() string {
	if key := os.Getenv("RESEND_API_KEY"); key != "" {
		return strings.TrimSpace(key)
	}
	if key := os.Getenv("SMTP_PASS"); strings.HasPrefix(key, "re_") {
		return strings.TrimSpace(key)
	}
	return ""
}
