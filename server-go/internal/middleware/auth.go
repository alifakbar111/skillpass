package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/uptrace/bun"

	"skillpass-server-go/internal/lib"
	"skillpass-server-go/internal/models"
)

type Claims struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// DefaultSSEExchangeTTL is how long an SSE exchange ticket is valid for.
// Tickets are also single-use: a successful Redeem deletes the entry.
const DefaultSSEExchangeTTL = 60 * time.Second

// StreamExchangeEntry is the value stored against an exchange ticket.
type StreamExchangeEntry struct {
	UserID    string
	Role      string
	ExpiresAt time.Time
}

// StreamExchangeStore holds short-lived, single-use SSE exchange tickets
// issued by the auth handler and redeemed by the SSE stream middleware.
// The in-memory map is protected by a RWMutex and a periodic GC sweeps
// expired entries.
type StreamExchangeStore struct {
	mu      sync.RWMutex
	entries map[string]StreamExchangeEntry
}

func NewStreamExchangeStore() *StreamExchangeStore {
	s := &StreamExchangeStore{entries: make(map[string]StreamExchangeEntry)}
	go s.gc()
	return s
}

// Issue generates a new opaque ticket bound to userID+role. The ticket
// expires after ttl and is consumed on first use.
func (s *StreamExchangeStore) Issue(userID, role string, ttl time.Duration) (string, error) {
	if userID == "" || role == "" {
		return "", fmt.Errorf("stream exchange: userID and role are required")
	}
	if ttl <= 0 {
		ttl = DefaultSSEExchangeTTL
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("stream exchange: read random: %w", err)
	}
	nonce := hex.EncodeToString(buf)
	s.mu.Lock()
	s.entries[nonce] = StreamExchangeEntry{
		UserID:    userID,
		Role:      role,
		ExpiresAt: time.Now().Add(ttl),
	}
	s.mu.Unlock()
	return nonce, nil
}

// Redeem looks up the ticket, deletes it on success, and returns the
// bound userID and role. Returns ok=false if the ticket is missing,
// already used, or expired.
func (s *StreamExchangeStore) Redeem(ticket string) (userID, role string, ok bool) {
	if ticket == "" {
		return "", "", false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, found := s.entries[ticket]
	if !found {
		return "", "", false
	}
	delete(s.entries, ticket)
	if time.Now().After(entry.ExpiresAt) {
		return "", "", false
	}
	return entry.UserID, entry.Role, true
}

// Size returns the number of live entries. Intended for tests.
func (s *StreamExchangeStore) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}

func (s *StreamExchangeStore) gc() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for k, e := range s.entries {
			if now.After(e.ExpiresAt) {
				delete(s.entries, k)
			}
		}
		s.mu.Unlock()
	}
}

// SSERedeemMiddleware authenticates an SSE stream request by redeeming
// an exchange ticket supplied as ?exchange=<nonce>. The ticket is
// single-use and short-lived; on success the userId and role are
// attached to the gin context for downstream handlers.
func SSERedeemMiddleware(store *StreamExchangeStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		ticket := c.Query("exchange")
		if ticket == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing exchange ticket"})
			return
		}
		userID, role, ok := store.Redeem(ticket)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired exchange ticket"})
			return
		}
		c.Set("userId", userID)
		c.Set("role", role)
		c.Next()
	}
}

// GetUserID extracts the authenticated userId from the gin context.
// Returns ("", false) if the key is missing, not a string, or empty.
// Prefer this over the comma-ok idiom in every handler.
func GetUserID(c *gin.Context) (string, bool) {
	v, ok := c.Get("userId")
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return "", false
	}
	return s, true
}

// GetRole extracts the authenticated role from the gin context.
// Returns ("", false) if the key is missing, not a string, or empty.
func GetRole(c *gin.Context) (string, bool) {
	v, ok := c.Get("role")
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return "", false
	}
	return s, true
}

func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		// EventSource cannot set custom headers cross-browser. The accepted
		// way to authenticate an SSE stream is now the exchange-ticket flow
		// (?exchange=<nonce>) handled by SSERedeemMiddleware on the stream
		// route itself. AuthRequired no longer accepts ?token= as a fallback
		// because tokens in URLs leak into server logs, browser history, and
		// Referer headers.
		if !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		token, err := jwt.ParseWithClaims(auth[7:], &Claims{}, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return []byte(jwtSecret), nil
		}, jwt.WithExpirationRequired(), jwt.WithValidMethods([]string{"HS256"}))
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		c.Set("userId", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists || userRole.(string) != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		c.Next()
	}
}

func RequireVerifiedCompany(bunDB *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("userId")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		userIDStr, ok := userIDVal.(string)
		if !ok || userIDStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userUUID, err := lib.ParseUUID(userIDStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid user ID: %v", err)})
			return
		}

		var companies []models.Company
		err = bunDB.NewSelect().
			Model(&companies).
			Column("id", "verification_status").
			Where("user_id = ?", userUUID).
			Scan(c.Request.Context())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Company lookup failed"})
			return
		}
		if len(companies) == 0 {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Company not found"})
			return
		}
		company := companies[0]
		if company.VerificationStatus != "verified" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Company not verified"})
			return
		}
		c.Set("companyId", company.ID.String())
		c.Next()
	}
}
