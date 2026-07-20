package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"golang.org/x/crypto/bcrypt"

	"skillpass-server-go/internal/authtoken"
	"skillpass-server-go/internal/email"
	"skillpass-server-go/internal/lib"
	"skillpass-server-go/internal/middleware"
	"skillpass-server-go/internal/models"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
	minPasswordLen  = 8
	refreshCookie   = "refreshToken"
)

// dummyBcryptHash is a real bcrypt hash of an unguessable random value,
// computed once at package init. Used to equalize the timing of
// ForgotPassword and any other "compare a secret" path so an attacker
// cannot distinguish "user exists" from "user does not exist" by
// measuring response time.
var (
	dummyBcryptHashOnce sync.Once
	dummyBcryptHash     string
)

func getDummyBcryptHash() string {
	dummyBcryptHashOnce.Do(func() {
		// A random 32-byte value that is never a real password.
		buf := make([]byte, 32)
		if _, err := rand.Read(buf); err != nil {
			// Fall back to a static hash if the RNG fails (extremely unlikely).
			dummyBcryptHash = "$2a$12$abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUV"
			return
		}
		hash, err := bcrypt.GenerateFromPassword(buf, 12)
		if err != nil {
			dummyBcryptHash = "$2a$12$abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUV"
			return
		}
		dummyBcryptHash = string(hash)
	})
	return dummyBcryptHash
}

// constantTimeEqualize runs a bcrypt compare against a dummy hash so the
// caller takes ~the same time whether or not the secret matches. The
// compare will always fail; the goal is only to equalize wall-clock.
func constantTimeEqualize(password string) {
	_ = bcrypt.CompareHashAndPassword([]byte(getDummyBcryptHash()), []byte(password))
}

// accessTokenCookie is the name of the HttpOnly cookie that mirrors the
// access token. The web client uses it via credentials: 'include', which
// removes the need to keep the token in localStorage (XSS-readable).
const accessTokenCookie = "accessToken"

// setAccessTokenCookie writes the access token as an HttpOnly,
// SameSite=Strict cookie. The Secure flag is set when GIN_MODE=release.
func setAccessTokenCookie(c *gin.Context, token string, ttl time.Duration) {
	secure := os.Getenv("GIN_MODE") == "release"
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     accessTokenCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(ttl.Seconds()),
	})
}

// clearAccessTokenCookie expires the access token cookie. Called on
// logout.
func clearAccessTokenCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     accessTokenCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   os.Getenv("GIN_MODE") == "release",
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
} //@name LoginRequest

type RegisterRequest struct {
	Email                string  `json:"email" binding:"required,email"`
	Username             string  `json:"username" binding:"required,min=3,max=32"`
	Password             string  `json:"password" binding:"required,min=8,max=128"`
	Name                 string  `json:"name"`
	Role                 string  `json:"role" binding:"required,oneof=jobseeker company"`
	CompanyName          *string `json:"companyName,omitempty"`
	BusinessRegistration *string `json:"businessRegistration,omitempty"`
	Website              *string `json:"website,omitempty"`
	Address              *string `json:"address,omitempty"`
	Contact              *string `json:"contact,omitempty"`
} //@name RegisterRequest

type UserResponse struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Username   string `json:"username"`
	Name       string `json:"name"`
	Role       string `json:"role"`
	IsVerified bool   `json:"isVerified"`
} //@name UserResponse

type LoginResponse struct {
	AccessToken string       `json:"accessToken"`
	User        UserResponse `json:"user"`
} //@name LoginResponse

type AuthHandler struct {
	db        *sql.DB
	bunDB     *bun.DB
	jwtSecret string
	emailer   email.Sender
	tokens    *authtoken.Service
	sseStore  *middleware.StreamExchangeStore
}

func NewAuthHandler(db *sql.DB, jwtSecret string, bunDB *bun.DB) *AuthHandler {
	return &AuthHandler{db: db, bunDB: bunDB, jwtSecret: jwtSecret}
}

// SetEmailer attaches an email sender for verification/reset mail.
// Optional — when nil, those emails are skipped.
func (h *AuthHandler) SetEmailer(e email.Sender) {
	h.emailer = e
}

// SetTokenService attaches the verification/reset token service.
// Optional — when nil, verification and reset flows are disabled.
func (h *AuthHandler) SetTokenService(t *authtoken.Service) {
	h.tokens = t
}

// SetSSEStore attaches the in-memory store for short-lived SSE exchange
// tickets. Required only if /auth/sse-ticket is exposed.
func (h *AuthHandler) SetSSEStore(s *middleware.StreamExchangeStore) {
	h.sseStore = s
}

type tokenClaims struct {
	UserID string
	Role   string
	Type   string
}

func (h *AuthHandler) signTokens(c *gin.Context, db bun.IDB, userID, role string) (accessToken, refreshToken string, refreshID uuid.UUID, err error) {
	now := time.Now()
	// Read the current token_version so admin-initiated role changes
	// (which bump the column) immediately invalidate outstanding tokens.
	var tokenVersion int
	if err = db.NewSelect().
		Model((*models.User)(nil)).
		Column("token_version").
		Where("id = ?", userID).
		Scan(c.Request.Context(), &tokenVersion); err != nil {
		return "", "", uuid.Nil, fmt.Errorf("lookup token version: %w", err)
	}

	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":       userID,
		"role":         role,
		"tokenVersion": tokenVersion,
		"iat":          now.Unix(),
		"exp":          now.Add(accessTokenTTL).Unix(),
	}).SignedString([]byte(h.jwtSecret))
	if err != nil {
		return "", "", uuid.Nil, err
	}

	refreshID = uuid.New()
	refreshExpires := now.Add(refreshTokenTTL)
	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"jti":          refreshID.String(),
		"userId":       userID,
		"role":         role,
		"tokenVersion": tokenVersion,
		"type":         "refresh",
		"iat":          now.Unix(),
		"exp":          refreshExpires.Unix(),
	}).SignedString([]byte(h.jwtSecret))
	if err != nil {
		return "", "", uuid.Nil, err
	}

	hash := hashToken(refreshToken)
	rt := &models.RefreshToken{
		ID:        refreshID,
		UserID:    uuid.MustParse(userID),
		TokenHash: hash,
		ExpiresAt: refreshExpires,
	}
	ctx := c.Request.Context()
	if _, err = db.NewInsert().Model(rt).Exec(ctx); err != nil {
		return "", "", uuid.Nil, err
	}

	setRefreshCookie(c, refreshToken, refreshTokenTTL)
	return accessToken, refreshToken, refreshID, nil
}

func setRefreshCookie(c *gin.Context, token string, ttl time.Duration) {
	secure := os.Getenv("COOKIE_SECURE") == "true" || os.Getenv("GIN_MODE") == "release"
	cookie := &http.Cookie{
		Name:     refreshCookie,
		Value:    token,
		Path:     "/api/v1/auth",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(ttl.Seconds()),
	}
	http.SetCookie(c.Writer, cookie)
}

func clearRefreshCookie(c *gin.Context) {
	cookie := &http.Cookie{
		Name:     refreshCookie,
		Value:    "",
		Path:     "/api/v1/auth",
		HttpOnly: true,
		MaxAge:   -1,
	}
	http.SetCookie(c.Writer, cookie)
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// SSETicketResponse	godoc
// @Description	Short-lived single-use exchange ticket for authenticating an SSE stream. Pass as ?exchange=<ticket> to the stream endpoint.
// @name			SSETicketResponse
type SSETicketResponse struct {
	Exchange  string `json:"exchange"`
	ExpiresIn int    `json:"expiresIn"`
}

// CreateSSETicket	godoc
// @Summary		Issue an SSE exchange ticket
// @Description	Exchanges a Bearer access token for a short-lived, single-use opaque ticket that can authenticate an EventSource stream. The ticket is bound to the caller's userId and expires after 60 seconds. Intended for clients that cannot set Authorization headers (e.g. browser EventSource).
// @Tags		auth
// @Produce		json
// @Security	BearerAuth
// @Success		200 {object} SSETicketResponse
// @Failure		401 {object} map[string]string
// @Router		/auth/sse-ticket [post]
func (h *AuthHandler) CreateSSETicket(c *gin.Context) {
	if h.sseStore == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SSE exchange store not configured"})
		return
	}
	userIDVal, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, ok := userIDVal.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	roleVal, ok := c.Get("role")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	role, ok := roleVal.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	nonce, err := h.sseStore.Issue(userID, role, middleware.DefaultSSEExchangeTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to issue exchange ticket"})
		return
	}
	c.JSON(http.StatusOK, SSETicketResponse{
		Exchange:  nonce,
		ExpiresIn: int(middleware.DefaultSSEExchangeTTL.Seconds()),
	})
}

// Register		godoc
// @Summary		Register a new user
// @Description	Create a new user account (jobseeker or company). Sets a refresh token cookie and returns an access token.
// @Tags		auth
// @Accept		json
// @Produce		json
// @Param		body body RegisterRequest true "Registration details"
// @Success		201 {object} LoginResponse
// @Failure		400 {object} map[string]string
// @Failure		409 {object} map[string]string
// @Router		/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	displayName := req.Name
	if req.Role == "company" {
		if req.CompanyName == nil || *req.CompanyName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		displayName = *req.CompanyName
	} else {
		if displayName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
	}

	passwordHash, err := lib.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	tx, err := h.bunDB.BeginTx(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
		return
	}
	defer tx.Rollback()

	user := &models.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: passwordHash,
		Name:         displayName,
		Role:         req.Role,
	}
	err = tx.NewInsert().Model(user).Returning("*").Scan(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Could not create account. The email already registered"})
		return
	}

	if req.Role == "jobseeker" {
		_, err = tx.NewInsert().Model(&models.JobseekerProfile{
			UserID: user.ID,
			Slug:   req.Username,
		}).Exec(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create profile"})
			return
		}
	} else {
		coName := displayName
		if req.CompanyName != nil {
			coName = *req.CompanyName
		}
		verificationDocs, _ := json.Marshal(map[string]string{
			"businessRegistration": coalesceStr(req.BusinessRegistration),
			"website":              coalesceStr(req.Website),
			"address":              coalesceStr(req.Address),
			"contact":              coalesceStr(req.Contact),
		})
		_, err = tx.NewInsert().Model(&models.Company{
			UserID:             user.ID,
			CompanyName:        coName,
			Website:            coalesceStrPtr(req.Website),
			Industry:           "Technology",
			VerificationDocs:   strPtr(string(verificationDocs)),
			VerificationStatus: "pending",
		}).Exec(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create company"})
			return
		}
	}

	accessToken, _, _, err := h.signTokens(c, tx, user.ID.String(), user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
		return
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit"})
		return
	}

	// Best-effort: send the welcome + email-verification mail.
	h.sendVerificationEmail(c.Request.Context(), user.ID.String(), user.Email, user.Name)

	setAccessTokenCookie(c, accessToken, accessTokenTTL)
	c.JSON(http.StatusCreated, LoginResponse{
		AccessToken: accessToken,
		User: UserResponse{
			ID:         user.ID.String(),
			Email:      user.Email,
			Username:   user.Username,
			Name:       user.Name,
			Role:       user.Role,
			IsVerified: false,
		},
	})
}

// Login		godoc
// @Summary		Login
// @Description	Authenticate with email and password. Sets a refresh token cookie and returns an access token.
// @Tags		auth
// @Accept		json
// @Produce		json
// @Param		body body LoginRequest true "Login credentials"
// @Success		200 {object} LoginResponse
// @Failure		401 {object} map[string]string
// @Router		/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var user models.User
	err := h.bunDB.NewSelect().Model(&user).
		Column("id", "email", "username", "name", "role", "password_hash", "is_verified").
		Where("email = ?", req.Email).
		Scan(c.Request.Context())
	if err != nil {
		_, _ = lib.HashPassword("dummy-equalize-timing")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	valid, err := lib.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	accessToken, _, _, err := h.signTokens(c, h.bunDB, user.ID.String(), user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
		return
	}

	// PR-22: also set the access token as an HttpOnly cookie so the web
	// client does not have to keep it in localStorage (XSS-readable).
	// SameSite=Strict blocks cross-site requests from sending the cookie.
	setAccessTokenCookie(c, accessToken, accessTokenTTL)

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken: accessToken,
		User: UserResponse{
			ID:         user.ID.String(),
			Email:      user.Email,
			Username:   user.Username,
			Name:       user.Name,
			Role:       user.Role,
			IsVerified: user.IsVerified,
		},
	})
}

// Refresh		godoc
// @Summary		Refresh access token
// @Description	Exchange a valid refresh token (from cookie) for a new access token. Rotates the refresh token.
// @Tags		auth
// @Produce		json
// @Success		200 {object} RefreshResponse
// @Failure		401 {object} map[string]string
// @Router		/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	cookie, err := c.Cookie(refreshCookie)
	if err != nil || cookie == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	claims, err := parseRefreshToken(cookie, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	tokenHash := hashToken(cookie)

	tx, err := h.bunDB.BeginTx(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
		return
	}
	defer tx.Rollback()

	var rt models.RefreshToken
	err = tx.NewSelect().Model(&rt).
		Where("token_hash = ?", tokenHash).
		For("UPDATE").
		Scan(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}
	if rt.RevokedAt != nil || time.Now().After(rt.ExpiresAt) {
		revokeAllForUser(c.Request.Context(), tx, rt.UserID)
		_ = tx.Commit()
		clearRefreshCookie(c)
	clearAccessTokenCookie(c)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	accessToken, _, _, err := h.signTokens(c, tx, rt.UserID.String(), claims.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
		return
	}

	_, err = tx.NewUpdate().Model((*models.RefreshToken)(nil)).
		Set("revoked_at = ?", time.Now()).
		Where("id = ?", rt.ID).
		Exec(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to rotate token"})
		return
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit"})
		return
	}

	setAccessTokenCookie(c, accessToken, accessTokenTTL)
	c.JSON(http.StatusOK, RefreshResponse{AccessToken: accessToken})
}

// Logout		godoc
// @Summary		Logout
// @Description	Revoke all refresh tokens for the authenticated user and clear the refresh cookie.
// @Tags		auth
// @Produce		json
// @Security	BearerAuth
// @Success		200 {object} MessageResponse
// @Router		/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userIDVal, exists := c.Get("userId")
	if !exists {
		clearRefreshCookie(c)
	clearAccessTokenCookie(c)
		c.JSON(http.StatusOK, MessageResponse{Message: "Logged out"})
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok {
		clearRefreshCookie(c)
	clearAccessTokenCookie(c)
		c.JSON(http.StatusOK, MessageResponse{Message: "Logged out"})
		return
	}

	if err := revokeAllForUserString(c.Request.Context(), h.bunDB, userIDStr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log out"})
		return
	}
	clearRefreshCookie(c)
	clearAccessTokenCookie(c)
	clearAccessTokenCookie(c)
	c.JSON(http.StatusOK, MessageResponse{Message: "Logged out"})
}

func parseRefreshToken(tokenStr, secret string) (*tokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.MapClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return []byte(secret), nil
	}, jwt.WithExpirationRequired(), jwt.WithValidMethods([]string{"HS256"}))
	if err != nil || !token.Valid {
		return nil, errInvalidToken
	}
	mc, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		return nil, errInvalidToken
	}
	if t, _ := (*mc)["type"].(string); t != "refresh" {
		return nil, errInvalidToken
	}
	userID, _ := (*mc)["userId"].(string)
	role, _ := (*mc)["role"].(string)
	if userID == "" {
		return nil, errInvalidToken
	}
	return &tokenClaims{UserID: userID, Role: role, Type: "refresh"}, nil
}

var errInvalidToken = errors.New("invalid refresh token")

func revokeAllForUserString(ctx context.Context, bunDB bun.IDB, userID string) error {
	uid, err := lib.ParseUUID(userID)
	if err != nil {
		return nil
	}
	_, err = bunDB.NewUpdate().
		Model((*models.RefreshToken)(nil)).
		Set("revoked_at = ?", time.Now()).
		Where("user_id = ? AND revoked_at IS NULL", uid).
		Exec(ctx)
	return err
}

func strPtr(s string) *string { return &s }

func coalesceStrPtr(s *string) *string {
	if s == nil {
		return strPtr("")
	}
	return s
}

func coalesceStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
