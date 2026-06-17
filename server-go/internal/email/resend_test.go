package email

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/resend/resend-go/v3"
)

// mockRoundTripper implements http.RoundTripper for testing.
type mockRoundTripper struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func TestResendAPIKey(t *testing.T) {
	t.Run("reads from RESEND_API_KEY", func(t *testing.T) {
		os.Setenv("RESEND_API_KEY", "re_abc123")
		defer os.Unsetenv("RESEND_API_KEY")
		os.Unsetenv("SMTP_PASS")
		if got := ResendAPIKey(); got != "re_abc123" {
			t.Fatalf("expected re_abc123, got %q", got)
		}
	})

	t.Run("falls back to SMTP_PASS when it starts with re_", func(t *testing.T) {
		os.Unsetenv("RESEND_API_KEY")
		os.Setenv("SMTP_PASS", "re_def456")
		defer os.Unsetenv("SMTP_PASS")
		if got := ResendAPIKey(); got != "re_def456" {
			t.Fatalf("expected re_def456, got %q", got)
		}
	})

	t.Run("returns empty when no resend key available", func(t *testing.T) {
		os.Unsetenv("RESEND_API_KEY")
		os.Setenv("SMTP_PASS", "smtp-password-not-resend")
		defer os.Unsetenv("SMTP_PASS")
		if got := ResendAPIKey(); got != "" {
			t.Fatalf("expected empty, got %q", got)
		}
	})
}

func TestNewSenderFactoryResend(t *testing.T) {
	t.Run("returns ResendSender when RESEND_API_KEY is set", func(t *testing.T) {
		os.Setenv("RESEND_API_KEY", "re_test")
		defer os.Unsetenv("RESEND_API_KEY")
		os.Unsetenv("SMTP_HOST")
		if _, ok := NewSender().(*ResendSender); !ok {
			t.Fatal("expected ResendSender when RESEND_API_KEY set")
		}
	})

	t.Run("returns SMTPSender when RESEND_API_KEY empty and SMTP_HOST set", func(t *testing.T) {
		os.Unsetenv("RESEND_API_KEY")
		os.Unsetenv("SMTP_PASS")
		os.Setenv("SMTP_HOST", "smtp.example.com")
		defer os.Unsetenv("SMTP_HOST")
		if _, ok := NewSender().(*SMTPSender); !ok {
			t.Fatal("expected SMTPSender when RESEND_API_KEY empty and SMTP_HOST set")
		}
	})

	t.Run("returns ConsoleSender when nothing is set", func(t *testing.T) {
		os.Unsetenv("RESEND_API_KEY")
		os.Unsetenv("SMTP_PASS")
		os.Unsetenv("SMTP_HOST")
		if _, ok := NewSender().(*ConsoleSender); !ok {
			t.Fatal("expected ConsoleSender when neither RESEND_API_KEY nor SMTP_HOST set")
		}
	})

	t.Run("ResendSender takes priority over SMTP", func(t *testing.T) {
		os.Setenv("RESEND_API_KEY", "re_test")
		defer os.Unsetenv("RESEND_API_KEY")
		os.Setenv("SMTP_HOST", "smtp.example.com")
		defer os.Unsetenv("SMTP_HOST")
		if _, ok := NewSender().(*ResendSender); !ok {
			t.Fatal("expected ResendSender when both RESEND_API_KEY and SMTP_HOST set")
		}
	})
}

func TestResendSenderSend(t *testing.T) {
	// Helper to create a ResendSender with a mock HTTP client.
	newSenderWithMock := func(t *testing.T, roundTripFunc func(*http.Request) (*http.Response, error)) *ResendSender {
		t.Helper()
		os.Setenv("EMAIL_FROM", "test@skillpass.com")
		t.Cleanup(func() { os.Unsetenv("EMAIL_FROM") })

		mock := &mockRoundTripper{roundTripFunc: roundTripFunc}
		httpClient := &http.Client{Transport: mock}
		client := resend.NewCustomClient(httpClient, "re_test")
		return &ResendSender{client: client, from: "test@skillpass.com"}
	}

	t.Run("success", func(t *testing.T) {
		s := newSenderWithMock(t, func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"id": "email_123"}`)),
				Header:     make(http.Header),
			}, nil
		})

		err := s.Send(context.Background(), "to@example.com", "Test", "<p>html</p>", "text")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("api error (non-200)", func(t *testing.T) {
		s := newSenderWithMock(t, func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader(`{"error": "invalid from address"}`)),
				Header:     make(http.Header),
			}, nil
		})

		err := s.Send(context.Background(), "to@example.com", "Test", "<p>html</p>", "text")
		if err == nil {
			t.Fatal("expected error but got nil")
		}
	})

	t.Run("network error", func(t *testing.T) {
		s := newSenderWithMock(t, func(req *http.Request) (*http.Response, error) {
			return nil, io.ErrUnexpectedEOF
		})

		err := s.Send(context.Background(), "to@example.com", "Test", "<p>html</p>", "text")
		if err == nil {
			t.Fatal("expected error but got nil")
		}
	})

	t.Run("sends correct request body", func(t *testing.T) {
		s := newSenderWithMock(t, func(req *http.Request) (*http.Response, error) {
			body, _ := io.ReadAll(req.Body)
			bodyStr := string(body)

			if !strings.Contains(bodyStr, `"to":["to@example.com"]`) {
				t.Errorf("request body missing recipient")
			}
			if !strings.Contains(bodyStr, `"from":"test@skillpass.com"`) {
				t.Errorf("request body missing from")
			}
			if !strings.Contains(bodyStr, `"subject":"Test"`) {
				t.Errorf("request body missing subject")
			}
			if !strings.Contains(bodyStr, `"html"`) {
				t.Errorf("request body missing html field, got: %s", bodyStr)
			}
			if !strings.Contains(bodyStr, `"text":"text"`) {
				t.Errorf("request body missing text")
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"id": "email_123"}`)),
				Header:     make(http.Header),
			}, nil
		})

		err := s.Send(context.Background(), "to@example.com", "Test", "<p>html</p>", "text")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestResendSenderWithContext(t *testing.T) {
	// Verify context cancellation is respected.
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	mock := &mockRoundTripper{roundTripFunc: func(req *http.Request) (*http.Response, error) {
		return nil, context.Canceled
	}}
	httpClient := &http.Client{Transport: mock}
	client := resend.NewCustomClient(httpClient, "re_test")
	s := &ResendSender{client: client, from: "test@skillpass.com"}

	err := s.Send(ctx, "to@example.com", "Test", "<p>html</p>", "text")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
