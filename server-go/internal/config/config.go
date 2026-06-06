package config

import "os"

type Config struct {
	Port        string
	JWTSecret   string
	DatabaseURL string
	CORSOrigin  string
}

func Load() *Config {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		panic("JWT_SECRET environment variable is required")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		panic("DATABASE_URL environment variable is required")
	}

	return &Config{
		Port:        getEnv("PORT", "1234"),
		JWTSecret:   jwtSecret,
		DatabaseURL: dbURL,
		CORSOrigin:  getEnv("CORS_ORIGIN", "http://localhost:4200"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
