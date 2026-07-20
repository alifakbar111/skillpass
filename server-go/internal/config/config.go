package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port          string
	JWTSecret     string
	DatabaseURL   string
	CORSOrigin    string
	ServeStatic   bool
	MarkItDownURL string
}

// MinJWTSecretLen is the minimum acceptable JWT_SECRET length in bytes.
// Anything shorter is refused at startup to prevent weak signing keys.
const MinJWTSecretLen = 32

// weakJWTSecrets is the deny-list of known-weak values. Any match causes
// Load() to panic at startup. The dev placeholder shipped in .env.example
// must always appear here.
var weakJWTSecrets = []string{
	"skillpass-dev-secret-change-in-prod",
	"password",
	"secret",
	"changeme",
	"change-me",
	"change_me",
}

func Load() *Config {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		panic("JWT_SECRET environment variable is required")
	}
	if len(jwtSecret) < MinJWTSecretLen {
		panic(fmt.Sprintf("JWT_SECRET must be at least %d bytes (got %d) — generate a strong one with `openssl rand -hex 64`", MinJWTSecretLen, len(jwtSecret)))
	}
	for _, weak := range weakJWTSecrets {
		if jwtSecret == weak {
			panic(fmt.Sprintf("JWT_SECRET is a known-weak value (%q) — generate a new one with `openssl rand -hex 64`", weak))
		}
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		panic("DATABASE_URL environment variable is required")
	}

	return &Config{
		Port:          getEnv("PORT", "1234"),
		JWTSecret:     jwtSecret,
		DatabaseURL:   dbURL,
		CORSOrigin:    getEnv("CORS_ORIGIN", "http://localhost:4200"),
		ServeStatic:   getEnv("SERVE_STATIC", "true") == "true",
		MarkItDownURL: getEnv("MARKITDOWN_URL", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
