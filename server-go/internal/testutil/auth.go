package testutil

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const TestJWTSecret = "test-secret-for-testing-only"

func GenerateToken(userID, role string, ttl time.Duration) string {
	now := time.Now()
	claims := jwt.MapClaims{
		"userId": userID,
		"role":   role,
		"iat":    now.Unix(),
		"exp":    now.Add(ttl).Unix(),
	}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(TestJWTSecret))
	return token
}

func GenerateExpiredToken(userID, role string) string {
	now := time.Now()
	claims := jwt.MapClaims{
		"userId": userID,
		"role":   role,
		"iat":    now.Add(-2 * time.Hour).Unix(),
		"exp":    now.Add(-1 * time.Hour).Unix(),
	}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(TestJWTSecret))
	return token
}

func GenerateRefreshToken(userID, role, jti string, ttl time.Duration) string {
	now := time.Now()
	claims := jwt.MapClaims{
		"jti":    jti,
		"userId": userID,
		"role":   role,
		"type":   "refresh",
		"iat":    now.Unix(),
		"exp":    now.Add(ttl).Unix(),
	}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(TestJWTSecret))
	return token
}
