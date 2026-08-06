package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port          string
	JWTSecret     string
	DatabaseURL   string
	CORSOrigin    string
	ServeStatic   bool
	MarkItDownURL string

	// Phase 2 · Sprint 1 — documents / storage / scanning.
	DocumentsDir    string // private dir for local document storage
	ClamAVAddr      string // host:port of clamd (INSTREAM); empty = scanning skipped
	StorageProvider string // "disk" (default) | "s3"
	RedisURL        string // for Asynq (next increment); empty = inline processing
	S3Bucket        string
	S3Region        string
	S3Endpoint      string
	S3AccessKey     string
	S3SecretKey     string

	// Phase 2 · Sprint 2 — face recognition.
	FaceServiceURL      string  // base URL of the Python face-service; empty = disabled
	FaceMatchThreshold  float64 // accept a verification at/above this match score
	FaceReviewThreshold float64 // flag for review between review and match thresholds
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

		DocumentsDir:    getEnv("DOCUMENTS_DIR", "./data/documents"),
		ClamAVAddr:      getEnv("CLAMAV_ADDR", ""),
		StorageProvider: getEnv("STORAGE_PROVIDER", "disk"),
		RedisURL:        getEnv("REDIS_URL", ""),
		S3Bucket:        getEnv("S3_BUCKET", ""),
		S3Region:        getEnv("S3_REGION", ""),
		S3Endpoint:      getEnv("S3_ENDPOINT", ""),
		S3AccessKey:     getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey:     getEnv("S3_SECRET_KEY", ""),

		FaceServiceURL:      getEnv("FACE_SERVICE_URL", ""),
		FaceMatchThreshold:  getEnvFloat("FACE_MATCH_THRESHOLD", 0.82),
		FaceReviewThreshold: getEnvFloat("FACE_REVIEW_THRESHOLD", 0.70),
	}
}

func getEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
