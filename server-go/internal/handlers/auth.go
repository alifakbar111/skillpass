package handlers

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"

	"skillpass-server-go/.gen/skillpass/public/model"
	"skillpass-server-go/internal/authtoken"
	"skillpass-server-go/internal/email"
	"skillpass-server-go/internal/gen"
	"skillpass-server-go/internal/lib"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
	minPasswordLen  = 8
	refreshCookie   = "refreshToken"
)

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
	CompanyName          *string `json:"companyName"`
	BusinessRegistration *string `json:"businessRegistration"`
	Website              *string `json:"website"`
	Address              *string `json:"address"`
	Contact              *string `json:"contact"`
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
	jwtSecret string
	emailer   email.Sender
	tokens    *authtoken.Service
}

func NewAuthHandler(db *sql.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{db: db, jwtSecret: jwtSecret}
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

type tokenClaims struct {
	UserID string
	Role   string
	Type   string
}

func (h *AuthHandler) signTokens(c *gin.Context, userID, role string) (accessToken, refreshToken string, refreshID uuid.UUID, err error) {
	now := time.Now()
	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userID,
		"role":   role,
		"iat":    now.Unix(),
		"exp":    now.Add(accessTokenTTL).Unix(),
	}).SignedString([]byte(h.jwtSecret))
	if err != nil {
		return "", "", uuid.Nil, err
	}

	refreshID = uuid.New()
	refreshExpires := now.Add(refreshTokenTTL)
	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"jti":   refreshID.String(),
		"userId": userID,
		"role":   role,
		"type":   "refresh",
		"iat":    now.Unix(),
		"exp":    refreshExpires.Unix(),
	}).SignedString([]byte(h.jwtSecret))
	if err != nil {
		return "", "", uuid.Nil, err
	}

	hash := hashToken(refreshToken)
	insertStmt := gen.RefreshTokens.INSERT(
		gen.RefreshTokens.ID, gen.RefreshTokens.UserID, gen.RefreshTokens.TokenHash,
		gen.RefreshTokens.ExpiresAt,
	).VALUES(
		refreshID, userID, hash, TimestampzT(refreshExpires),
	)
	ctx := c.Request.Context()
	if _, err = insertStmt.ExecContext(ctx, h.db); err != nil {
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

	tx, err := h.db.BeginTx(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
		return
	}
	defer tx.Rollback()

	var user model.Users
	insertStmt := gen.Users.INSERT(
		gen.Users.Email, gen.Users.Username, gen.Users.PasswordHash, gen.Users.Name, gen.Users.Role,
	).VALUES(
		req.Email, req.Username, passwordHash, displayName, req.Role,
	).RETURNING(
		gen.Users.ID, gen.Users.Email, gen.Users.Username, gen.Users.Name, gen.Users.Role,
	)

	if err = insertStmt.QueryContext(c.Request.Context(), tx, &user); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Could not create account"})
		return
	}

	if req.Role == "jobseeker" {
		_, err = gen.JobseekerProfiles.INSERT(
			gen.JobseekerProfiles.UserID, gen.JobseekerProfiles.Slug,
		).VALUES(
			user.ID, req.Username,
		).ExecContext(c.Request.Context(), tx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create profile"})
			return
		}
	} else {
		coName := displayName
		if req.CompanyName != nil {
			coName = *req.CompanyName
		}
		_, err = gen.Companies.INSERT(
			gen.Companies.UserID, gen.Companies.CompanyName, gen.Companies.Industry,
			gen.Companies.VerificationDocs, gen.Companies.VerificationStatus,
		).VALUES(
			user.ID, coName, "Technology",
			`{"businessRegistration":"","website":"","address":"","contact":""}`, gen.VerificationStatusPending,
		).ExecContext(c.Request.Context(), tx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create company"})
			return
		}
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit"})
		return
	}

	accessToken, _, _, err := h.signTokens(c, user.ID.String(), string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
		return
	}

	// Best-effort: send the welcome + email-verification mail.
	h.sendVerificationEmail(c.Request.Context(), user.ID.String(), user.Email, user.Name)

	c.JSON(http.StatusCreated, LoginResponse{
		AccessToken: accessToken,
		User: UserResponse{
			ID:         user.ID.String(),
			Email:      user.Email,
			Username:   user.Username,
			Name:       user.Name,
			Role:       string(user.Role),
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

	stmt := SELECT(
		gen.Users.ID, gen.Users.Email, gen.Users.Username, gen.Users.Name,
		gen.Users.Role, gen.Users.PasswordHash, gen.Users.IsVerified,
	).FROM(
		gen.Users,
	).WHERE(
		gen.Users.Email.EQ(String(req.Email)),
	)

	var user model.Users
	err := stmt.QueryContext(c.Request.Context(), h.db, &user)
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

	accessToken, _, _, err := h.signTokens(c, user.ID.String(), string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken: accessToken,
		User: UserResponse{
			ID:         user.ID.String(),
			Email:      user.Email,
			Username:   user.Username,
			Name:       user.Name,
			Role:       string(user.Role),
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

	tx, err := h.db.BeginTx(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
		return
	}
	defer tx.Rollback()

	rtStmt := SELECT(
		gen.RefreshTokens.ID, gen.RefreshTokens.UserID, gen.RefreshTokens.ExpiresAt,
		gen.RefreshTokens.RevokedAt, gen.RefreshTokens.ReplacedBy,
	).FROM(
		gen.RefreshTokens,
	).WHERE(
		gen.RefreshTokens.TokenHash.EQ(String(tokenHash)),
	).FOR(UPDATE())

	var rt model.RefreshTokens
	if err = rtStmt.QueryContext(c.Request.Context(), tx, &rt); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}
	if rt.RevokedAt != nil || time.Now().After(rt.ExpiresAt) {
		revokeAllForUser(c.Request.Context(), tx, rt.UserID)
		_ = tx.Commit()
		clearRefreshCookie(c)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	accessToken, _, _, err := h.signTokens(c, rt.UserID.String(), claims.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
		return
	}

	revokeStmt := gen.RefreshTokens.UPDATE().SET(
		gen.RefreshTokens.RevokedAt.SET(TimestampzT(time.Now())),
	).WHERE(
		gen.RefreshTokens.ID.EQ(UUID(rt.ID)),
	)
	if _, err = revokeStmt.ExecContext(c.Request.Context(), tx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to rotate token"})
		return
	}

	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit"})
		return
	}

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
		c.JSON(http.StatusOK, MessageResponse{Message: "Logged out"})
		return
	}
	userIDStr, ok := userIDVal.(string)
	if !ok {
		clearRefreshCookie(c)
		c.JSON(http.StatusOK, MessageResponse{Message: "Logged out"})
		return
	}

	if err := revokeAllForUserString(c.Request.Context(), h.db, userIDStr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to log out"})
		return
	}
	clearRefreshCookie(c)
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

func revokeAllForUserString(ctx context.Context, db *sql.DB, userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil
	}
	stmt := gen.RefreshTokens.UPDATE().SET(
		gen.RefreshTokens.RevokedAt.SET(TimestampzT(time.Now())),
	).WHERE(
		gen.RefreshTokens.UserID.EQ(UUID(uid)).AND(
			gen.RefreshTokens.RevokedAt.IS_NULL(),
		),
	)
	_, err = stmt.ExecContext(ctx, db)
	return err
}
