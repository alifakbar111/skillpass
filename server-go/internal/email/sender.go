// Package email delivers transactional email. It mirrors the lib.LLMClient
// pattern: a small interface, env-driven implementations, and a factory.
package email

import (
	"context"
	"fmt"
	"log/slog"
	"mime"
	"net/smtp"
	"os"
	"strings"
)

// Sender delivers a single transactional email. Implementations must be safe
// for concurrent use. Callers treat delivery as best-effort: log failures,
// never block the main operation on them.
type Sender interface {
	Send(ctx context.Context, to, subject, htmlBody, textBody string) error
}

// Compile-time interface checks.
var _ Sender = (*SMTPSender)(nil)
var _ Sender = (*ConsoleSender)(nil)

// NewSender creates a Sender based on the environment:
// if SMTP_HOST is set an SMTPSender is returned, otherwise a ConsoleSender
// that logs emails to stdout (dev default — no SMTP account needed).
func NewSender() Sender {
	if os.Getenv("SMTP_HOST") != "" {
		return NewSMTPSender()
	}
	return &ConsoleSender{}
}

// AppBaseURL returns the public web app URL used for links inside emails.
// Configured via APP_URL; defaults to the local dev web server.
func AppBaseURL() string {
	url := os.Getenv("APP_URL")
	if url == "" {
		url = "http://localhost:4200"
	}
	return strings.TrimRight(url, "/")
}

// SMTPSender sends mail through an SMTP relay. Env:
// SMTP_HOST (required), SMTP_PORT (default 587), SMTP_USER, SMTP_PASS,
// EMAIL_FROM (default no-reply@skillpass.local).
type SMTPSender struct {
	host string
	port string
	user string
	pass string
	from string
}

func NewSMTPSender() *SMTPSender {
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}
	from := os.Getenv("EMAIL_FROM")
	if from == "" {
		from = "no-reply@skillpass.local"
	}
	return &SMTPSender{
		host: os.Getenv("SMTP_HOST"),
		port: port,
		user: os.Getenv("SMTP_USER"),
		pass: os.Getenv("SMTP_PASS"),
		from: from,
	}
}

func (s *SMTPSender) Send(_ context.Context, to, subject, htmlBody, textBody string) error {
	msg := buildMIMEMessage(s.from, to, subject, htmlBody, textBody)

	addr := s.host + ":" + s.port
	var auth smtp.Auth
	if s.user != "" {
		auth = smtp.PlainAuth("", s.user, s.pass, s.host)
	}
	if err := smtp.SendMail(addr, auth, s.from, []string{to}, msg); err != nil {
		return fmt.Errorf("smtp send to %s: %w", to, err)
	}
	return nil
}

// buildMIMEMessage assembles a multipart/alternative message with plain-text
// and HTML parts so any mail client can render it.
func buildMIMEMessage(from, to, subject, htmlBody, textBody string) []byte {
	const boundary = "skillpass-mime-boundary"
	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: " + mime.QEncoding.Encode("utf-8", subject) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: multipart/alternative; boundary=" + boundary + "\r\n")
	b.WriteString("\r\n")

	b.WriteString("--" + boundary + "\r\n")
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
	b.WriteString(textBody + "\r\n")

	b.WriteString("--" + boundary + "\r\n")
	b.WriteString("Content-Type: text/html; charset=utf-8\r\n\r\n")
	b.WriteString(htmlBody + "\r\n")

	b.WriteString("--" + boundary + "--\r\n")
	return []byte(b.String())
}

// ConsoleSender logs the email to stdout instead of delivering it.
// Dev default so flows that send email (verification, reset links) work
// locally — the link is readable straight from the server log.
type ConsoleSender struct{}

func (c *ConsoleSender) Send(_ context.Context, to, subject, _, textBody string) error {
	slog.Info("email (console sender — not delivered)", "to", to, "subject", subject)
	fmt.Printf("\n────── EMAIL (console sender) ──────\nTo: %s\nSubject: %s\n\n%s\n─────────────────────────────────────\n\n", to, subject, textBody)
	return nil
}
