package middleware

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

// verifiedCompanyCacheTTL is how long a successful (userID → companyID,
// verified) lookup is cached. Short enough that revocation propagates
// within a few seconds, long enough to cut most repeated lookups.
const verifiedCompanyCacheTTL = 30 * time.Second

type verifiedCompanyEntry struct {
	companyID string
	verified  bool
	expiresAt time.Time
}

var (
	verifiedCompanyCacheMu sync.RWMutex
	verifiedCompanyCache   = make(map[string]verifiedCompanyEntry)
)

// invalidateVerifiedCompany removes a cached entry. Call this from
// admin/revoke paths so a revoked company cannot keep API access for
// the remainder of the TTL.
func invalidateVerifiedCompany(userID string) {
	verifiedCompanyCacheMu.Lock()
	delete(verifiedCompanyCache, userID)
	verifiedCompanyCacheMu.Unlock()
}

// lookupVerifiedCompany checks the cache first, then falls back to the
// DB. Returns (companyID, isVerified, ok). ok=false means the user has
// no company at all.
func lookupVerifiedCompany(ctx context.Context, bunDB *bun.DB, userID string) (string, bool, bool) {
	verifiedCompanyCacheMu.RLock()
	entry, hit := verifiedCompanyCache[userID]
	verifiedCompanyCacheMu.RUnlock()
	if hit && time.Now().Before(entry.expiresAt) {
		return entry.companyID, entry.verified, true
	}

	var companies []models.Company
	err := bunDB.NewSelect().
		Model(&companies).
		Column("id", "verification_status").
		Where("user_id = ?", userID).
		Scan(ctx)
	if err != nil {
		return "", false, false
	}
	if len(companies) == 0 {
		// Cache the "no company" result too, so repeated lookups don't
		// hit the DB. The error is handled by RequireVerifiedCompany.
		verifiedCompanyCacheMu.Lock()
		verifiedCompanyCache[userID] = verifiedCompanyEntry{
			companyID: "",
			verified:  false,
			expiresAt: time.Now().Add(verifiedCompanyCacheTTL),
		}
		verifiedCompanyCacheMu.Unlock()
		return "", false, false
	}
	company := companies[0]
	verified := company.VerificationStatus == "verified"
	verifiedCompanyCacheMu.Lock()
	verifiedCompanyCache[userID] = verifiedCompanyEntry{
		companyID: company.ID.String(),
		verified:  verified,
		expiresAt: time.Now().Add(verifiedCompanyCacheTTL),
	}
	verifiedCompanyCacheMu.Unlock()
	return company.ID.String(), verified, true
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
		userIDStr, ok := GetUserID(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userUUID, err := lib.ParseUUID(userIDStr)
		if err != nil {
			slog.Warn("invalid user ID in context", "error", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		companyID, verified, ok := lookupVerifiedCompany(c.Request.Context(), bunDB, userUUID.String())
		if !ok {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "Company not found"})
			return
		}

		// Multi-company support: a user belonging to multiple companies
		// may scope a request to one via ?companyId=. The override is
		// always subject to verification.
		if requestedID := c.Query("companyId"); requestedID != "" {
			if _, err := uuid.Parse(requestedID); err != nil {
				slog.Warn("invalid companyId query param", "raw", requestedID, "error", err)
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid companyId"})
				return
			}
			// Override requires a fresh DB check (the cache holds only the
			// default company; cross-company lookups are rare).
			overrideVerified, err := isCompanyVerified(c.Request.Context(), bunDB, requestedID)
			if err != nil {
				slog.Error("company override lookup failed", "companyId", requestedID, "error", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Company lookup failed"})
				return
			}
			if !overrideVerified {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Company not verified"})
				return
			}
			companyID = requestedID
		} else if !verified {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Company not verified"})
			return
		}

		c.Set("companyId", companyID)
		c.Next()
	}
}

// isCompanyVerified checks a specific company by ID, bypassing the
// default-company cache. Used for ?companyId= overrides.
func isCompanyVerified(ctx context.Context, bunDB *bun.DB, companyID string) (bool, error) {
	var status string
	err := bunDB.NewSelect().
		Model((*models.Company)(nil)).
		Column("verification_status").
		Where("id = ?", companyID).
		Scan(ctx, &status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return status == "verified", nil
}
