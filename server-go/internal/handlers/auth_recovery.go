package handlers

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"skillpass-server-go/internal/authtoken"
	"skillpass-server-go/internal/email"
	"skillpass-server-go/internal/lib"
)

// sendVerificationEmail issues a verification token and emails the link.
// Best-effort: failures are logged, never returned (registration must not
// fail because mail did).
func (h *AuthHandler) sendVerificationEmail(ctx context.Context, userID, to, name string) {
	if h.emailer == nil || h.tokens == nil {
		return
	}
	raw, err := h.tokens.CreateEmailVerification(ctx, userID)
	if err != nil {
		slog.Warn("create verification token failed", "userID", userID, "error", err)
		return
	}
	verifyURL := email.AppBaseURL() + "/auth/verify-email?token=" + raw
	msg := email.VerificationEmail(name, verifyURL)
	if err := h.emailer.Send(ctx, to, msg.Subject, msg.HTML, msg.Text); err != nil {
		slog.Warn("verification email delivery failed", "to", to, "error", err)
	}
}

// Me	godoc
// @Summary		Current user
// @Description	Return the authenticated user's account info (works for all roles, unlike /profiles/me which is jobseeker-only).
// @Tags		auth
// @Produce		json
// @Security	BearerAuth
// @Success		200 {object} UserResponse
// @Failure		401 {object} map[string]string
// @Router		/auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userIDVal, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, _ := userIDVal.(string)

	var resp UserResponse
	err := h.db.QueryRowContext(c.Request.Context(),
		`SELECT id, email, username, name, role::text, is_verified FROM users WHERE id = $1`,
		userID,
	).Scan(&resp.ID, &resp.Email, &resp.Username, &resp.Name, &resp.Role, &resp.IsVerified)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// VerifyEmail	godoc
// @Summary		Verify email address
// @Description	Consume an email-verification token (from the emailed link) and mark the account verified.
// @Tags		auth
// @Produce		json
// @Param		token query string true "Verification token"
// @Success		200 {object} map[string]string
// @Failure		400 {object} map[string]string
// @Router		/auth/verify-email [get]
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	if h.tokens == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Verification is not enabled"})
		return
	}
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing token"})
		return
	}

	if _, err := h.tokens.ConsumeEmailVerification(c.Request.Context(), token); err != nil {
		if errors.Is(err, authtoken.ErrInvalidToken) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "This verification link is invalid or has expired"})
			return
		}
		slog.Error("verify email failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Verification failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Email verified"})
}

// ResendVerification	godoc
// @Summary		Resend verification email
// @Description	Send a fresh verification link to the authenticated user's email.
// @Tags		auth
// @Produce		json
// @Security	BearerAuth
// @Success		200 {object} map[string]string
// @Failure		401 {object} map[string]string
// @Router		/auth/resend-verification [post]
func (h *AuthHandler) ResendVerification(c *gin.Context) {
	userIDVal, ok := c.Get("userId")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID, _ := userIDVal.(string)

	var to, name string
	var verified bool
	err := h.db.QueryRowContext(c.Request.Context(),
		`SELECT email, name, is_verified FROM users WHERE id = $1`, userID,
	).Scan(&to, &name, &verified)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	if verified {
		c.JSON(http.StatusOK, gin.H{"message": "Email already verified"})
		return
	}

	h.sendVerificationEmail(c.Request.Context(), userID, to, name)
	c.JSON(http.StatusOK, gin.H{"message": "Verification email sent"})
}

// ForgotPassword	godoc
// @Summary		Request password reset
// @Description	Send a password reset link if the email belongs to an account. Always responds with the same message to prevent account enumeration.
// @Tags		auth
// @Accept		json
// @Produce		json
// @Param		body body object{email=string} true "Account email"
// @Success		200 {object} map[string]string
// @Failure		400 {object} map[string]string
// @Router		/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "A valid email is required"})
		return
	}

	const genericMsg = "If that email belongs to an account, a reset link has been sent."

	if h.tokens == nil || h.emailer == nil {
		c.JSON(http.StatusOK, gin.H{"message": genericMsg})
		return
	}

	raw, user, err := h.tokens.CreatePasswordReset(c.Request.Context(), req.Email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			slog.Error("create password reset failed", "error", err)
		}
		// Same response whether or not the account exists.
		c.JSON(http.StatusOK, gin.H{"message": genericMsg})
		return
	}

	resetURL := email.AppBaseURL() + "/auth/reset-password?token=" + raw
	msg := email.PasswordResetEmail(user.Name, resetURL)
	if err := h.emailer.Send(c.Request.Context(), user.Email, msg.Subject, msg.HTML, msg.Text); err != nil {
		slog.Warn("password reset email delivery failed", "to", user.Email, "error", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": genericMsg})
}

// ResetPassword	godoc
// @Summary		Reset password
// @Description	Set a new password using a reset token from the emailed link. Revokes all existing sessions.
// @Tags		auth
// @Accept		json
// @Produce		json
// @Param		body body object{token=string,newPassword=string} true "Reset token and new password"
// @Success		200 {object} map[string]string
// @Failure		400 {object} map[string]string
// @Router		/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	if h.tokens == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Password reset is not enabled"})
		return
	}

	var req struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required,min=8,max=128"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token and a password of at least 8 characters are required"})
		return
	}

	hash, err := lib.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set password"})
		return
	}

	if _, err := h.tokens.ConsumePasswordReset(c.Request.Context(), req.Token, hash); err != nil {
		if errors.Is(err, authtoken.ErrInvalidToken) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "This reset link is invalid or has expired"})
			return
		}
		slog.Error("reset password failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated. Please sign in with your new password."})
}
