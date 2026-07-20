package config

import (
	"strings"
	"testing"
)

const validSecret = "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"

func TestLoad_RejectsEmptyJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	t.Setenv("DATABASE_URL", "postgres://localhost/x")

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected Load() to panic with empty JWT_SECRET")
		}
	}()
	_ = Load()
}

func TestLoad_RejectsShortJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "short")
	t.Setenv("DATABASE_URL", "postgres://localhost/x")

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected Load() to panic with short JWT_SECRET")
		}
	}()
	_ = Load()
}

func TestLoad_RejectsKnownDevJWTSecret(t *testing.T) {
	for _, weak := range weakJWTSecrets {
		t.Run(weak, func(t *testing.T) {
			t.Setenv("JWT_SECRET", weak)
			t.Setenv("DATABASE_URL", "postgres://localhost/x")

			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("expected Load() to panic with known-weak JWT_SECRET %q", weak)
				}
			}()
			_ = Load()
		})
	}
}

func TestLoad_AcceptsValidJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", validSecret)
	t.Setenv("DATABASE_URL", "postgres://localhost/x")

	cfg := Load()
	if cfg == nil {
		t.Fatal("expected Load() to return a non-nil *Config with a valid JWT_SECRET")
	}
	if cfg.JWTSecret != validSecret {
		t.Fatalf("expected JWTSecret to be %q, got %q", validSecret, cfg.JWTSecret)
	}
}

func TestLoad_RejectsJWTSecretJustBelow32Bytes(t *testing.T) {
	// 31 ASCII bytes — one short of the 32-byte minimum.
	t.Setenv("JWT_SECRET", strings.Repeat("a", 31))
	t.Setenv("DATABASE_URL", "postgres://localhost/x")

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected Load() to panic with 31-byte JWT_SECRET")
		}
	}()
	_ = Load()
}

func TestLoad_Accepts32ByteJWTSecret(t *testing.T) {
	// 32 ASCII bytes — the minimum allowed length.
	t.Setenv("JWT_SECRET", strings.Repeat("a", 32))
	t.Setenv("DATABASE_URL", "postgres://localhost/x")

	cfg := Load()
	if cfg == nil || cfg.JWTSecret == "" {
		t.Fatal("expected Load() to accept a 32-byte JWT_SECRET")
	}
}

func TestLoad_StillRejectsEmptyDatabaseURL(t *testing.T) {
	// Regression check: a valid JWT_SECRET must not mask the DATABASE_URL guard.
	t.Setenv("JWT_SECRET", validSecret)
	t.Setenv("DATABASE_URL", "")

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected Load() to panic with empty DATABASE_URL")
		}
	}()
	_ = Load()
}

func TestWeakSecretListCoversDevPlaceholder(t *testing.T) {
	// Hard guardrail: the example value shipped in .env.example must be in
	// the weak-secrets list, or the whole defense is theatre.
	for _, w := range weakJWTSecrets {
		if w == "skillpass-dev-secret-change-in-prod" {
			return
		}
	}
	t.Fatal("weakJWTSecrets list must include the dev placeholder \"skillpass-dev-secret-change-in-prod\"")
}
