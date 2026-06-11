package email

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestNewSenderFactory(t *testing.T) {
	t.Run("defaults to console sender", func(t *testing.T) {
		os.Unsetenv("SMTP_HOST")
		if _, ok := NewSender().(*ConsoleSender); !ok {
			t.Fatal("expected ConsoleSender when SMTP_HOST unset")
		}
	})

	t.Run("smtp when SMTP_HOST set", func(t *testing.T) {
		os.Setenv("SMTP_HOST", "smtp.example.com")
		defer os.Unsetenv("SMTP_HOST")
		if _, ok := NewSender().(*SMTPSender); !ok {
			t.Fatal("expected SMTPSender when SMTP_HOST set")
		}
	})
}

func TestConsoleSenderSend(t *testing.T) {
	c := &ConsoleSender{}
	if err := c.Send(context.Background(), "to@example.com", "Subj", "<p>html</p>", "text"); err != nil {
		t.Fatalf("console send: %v", err)
	}
}

func TestTemplates(t *testing.T) {
	t.Run("verification email", func(t *testing.T) {
		msg := VerificationEmail("Ada", "https://app.example.com/auth/verify-email?token=abc")
		if !strings.Contains(msg.Subject, "verify") {
			t.Fatalf("unexpected subject %q", msg.Subject)
		}
		if !strings.Contains(msg.HTML, "Ada") || !strings.Contains(msg.HTML, "token=abc") {
			t.Fatal("html missing name or link")
		}
		if !strings.Contains(msg.Text, "token=abc") {
			t.Fatal("text missing link")
		}
	})

	t.Run("escapes html in user values", func(t *testing.T) {
		msg := ApplicationReceivedEmail("Engineer", "<script>alert(1)</script>", "https://x")
		if strings.Contains(msg.HTML, "<script>") {
			t.Fatal("html not escaped")
		}
	})

	t.Run("status email mentions status", func(t *testing.T) {
		msg := StatusUpdateEmail("Engineer", "interviewed", "https://x")
		if !strings.Contains(msg.Text, "interviewed") {
			t.Fatal("text missing status")
		}
	})
}

func TestBuildMIMEMessage(t *testing.T) {
	msg := string(buildMIMEMessage("from@x.com", "to@y.com", "Hello", "<b>hi</b>", "hi"))
	for _, want := range []string{"From: from@x.com", "To: to@y.com", "multipart/alternative", "text/plain", "text/html"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("mime message missing %q", want)
		}
	}
}

func TestAppBaseURL(t *testing.T) {
	os.Unsetenv("APP_URL")
	if AppBaseURL() != "http://localhost:4200" {
		t.Fatalf("unexpected default %q", AppBaseURL())
	}
	os.Setenv("APP_URL", "https://skillpass.app/")
	defer os.Unsetenv("APP_URL")
	if AppBaseURL() != "https://skillpass.app" {
		t.Fatalf("expected trailing slash trimmed, got %q", AppBaseURL())
	}
}
