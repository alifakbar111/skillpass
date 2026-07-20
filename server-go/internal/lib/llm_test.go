package lib

import (
	"strings"
	"testing"
)

func TestPreviewBody(t *testing.T) {
	t.Run("short body returns trimmed string with hash", func(t *testing.T) {
		got := previewBody([]byte("hello world"), 200)
		if !strings.HasPrefix(got, "hello world") {
			t.Errorf("expected prefix %q, got %q", "hello world", got)
		}
		if !strings.Contains(got, "sha256=") {
			t.Errorf("expected sha256= in output, got %q", got)
		}
		if strings.Contains(got, "\n") {
			t.Errorf("output should not contain newlines, got %q", got)
		}
	})
	t.Run("long body is truncated", func(t *testing.T) {
		body := make([]byte, 500)
		for i := range body {
			body[i] = 'x'
		}
		got := previewBody(body, 50)
		if !strings.HasPrefix(got, strings.Repeat("x", 50)) {
			t.Errorf("expected 50-byte prefix, got %q", got[:60])
		}
		if !strings.Contains(got, "…") {
			t.Errorf("expected truncation marker, got %q", got)
		}
	})
	t.Run("identical bodies produce identical hashes", func(t *testing.T) {
		a := previewBody([]byte("same body"), 200)
		b := previewBody([]byte("same body"), 200)
		if a != b {
			t.Errorf("same input should produce same output: %q vs %q", a, b)
		}
	})
}
