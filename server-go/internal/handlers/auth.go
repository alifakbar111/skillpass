package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	. "github.com/go-jet/jet/v2/postgres"

	"skillpass-server-go/.gen/skillpass/public/model"
	"skillpass-server-go/internal/gen"
	"skillpass-server-go/internal/lib"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

type RegisterRequest struct {
	Email                string  `json:"email" binding:"required"`
	Username             string  `json:"username" binding:"required"`
	Password             string  `json:"password" binding:"required"`
	Name                 string  `json:"name" binding:"required"`
	Role                 string  `json:"role" binding:"required,oneof=jobseeker company"`
	CompanyName          *string `json:"companyName"`
	BusinessRegistration *string `json:"businessRegistration"`
	Website              *string `json:"website"`
	Address              *string `json:"address"`
	Contact              *string `json:"contact"`
}

type UserResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

type LoginResponse struct {
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
	User         UserResponse `json:"user"`
}

type AuthHandler struct {
	db        *sql.DB
	jwtSecret string
}

func NewAuthHandler(db *sql.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{db: db, jwtSecret: jwtSecret}
}

func (h *AuthHandler) signTokens(userID, role string) (accessToken, refreshToken string, err error) {
	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userID,
		"role":   role,
	}).SignedString([]byte(h.jwtSecret))
	if err != nil {
		return "", "", err
	}

	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userID,
		"role":   role,
		"type":   "refresh",
		"exp":    time.Now().Add(7 * 24 * time.Hour).Unix(),
	}).SignedString([]byte(h.jwtSecret))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var count int64
	countStmt := SELECT(
		COUNT(STAR).AS("count"),
	).FROM(
		gen.Users,
	).WHERE(
		gen.Users.Email.EQ(String(req.Email)),
	)

	var countResult struct{ Count int64 }
	err := countStmt.QueryContext(c.Request.Context(), h.db, &countResult)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	count = countResult.Count
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	displayName := req.Name
	if req.Role == "company" && req.CompanyName != nil && *req.CompanyName != "" {
		displayName = *req.CompanyName
	}

	passwordHash, err := lib.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	insertStmt := gen.Users.INSERT(
		gen.Users.Email, gen.Users.Username, gen.Users.PasswordHash, gen.Users.Name, gen.Users.Role,
	).VALUES(
		req.Email, req.Username, passwordHash, displayName, req.Role,
	).RETURNING(
		gen.Users.ID, gen.Users.Email, gen.Users.Username, gen.Users.Name, gen.Users.Role,
	)

	var user model.Users
	err = insertStmt.QueryContext(c.Request.Context(), h.db, &user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	if req.Role == "jobseeker" {
		_, err = gen.JobseekerProfiles.INSERT(
			gen.JobseekerProfiles.UserID, gen.JobseekerProfiles.Slug,
		).VALUES(
			user.ID, req.Username,
		).ExecContext(c.Request.Context(), h.db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create profile"})
			return
		}
	} else if req.Role == "company" {
		coName := displayName
		if req.CompanyName != nil {
			coName = *req.CompanyName
		}
		_, err = gen.Companies.INSERT(
			gen.Companies.UserID, gen.Companies.CompanyName, gen.Companies.Industry,
			gen.Companies.VerificationDocs, gen.Companies.VerificationStatus,
		).VALUES(
			user.ID, coName, "Technology",
			`{"businessRegistration":"","website":"","address":"","contact":""}`, "pending",
		).ExecContext(c.Request.Context(), h.db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create company"})
			return
		}
	}

	accessToken, refreshToken, err := h.signTokens(user.ID.String(), string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
		return
	}

	c.JSON(http.StatusCreated, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: UserResponse{
			ID:       user.ID.String(),
			Email:    user.Email,
			Username: user.Username,
			Name:     user.Name,
			Role:     string(user.Role),
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	stmt := SELECT(
		gen.Users.ID, gen.Users.Email, gen.Users.Username, gen.Users.Name,
		gen.Users.Role, gen.Users.PasswordHash,
	).FROM(
		gen.Users,
	).WHERE(
		gen.Users.Email.EQ(String(req.Email)),
	)

	var user model.Users
	err := stmt.QueryContext(c.Request.Context(), h.db, &user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	valid, err := lib.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	accessToken, refreshToken, err := h.signTokens(user.ID.String(), string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: UserResponse{
			ID:       user.ID.String(),
			Email:    user.Email,
			Username: user.Username,
			Name:     user.Name,
			Role:     string(user.Role),
		},
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	token, err := jwt.ParseWithClaims(req.RefreshToken, &jwt.MapClaims{}, func(t *jwt.Token) (any, error) {
		return []byte(h.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok || (*claims)["type"] != "refresh" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	userID := (*claims)["userId"].(string)
	role, ok := (*claims)["role"].(string)
	if !ok || role == "" {
		stmt := SELECT(
			gen.Users.Role,
		).FROM(
			gen.Users,
		).WHERE(
			gen.Users.ID.EQ(String(userID)),
		)

		var user model.Users
		err := stmt.QueryContext(c.Request.Context(), h.db, &user)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}
		role = string(user.Role)
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userID,
		"role":   role,
	}).SignedString([]byte(h.jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"accessToken": accessToken})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}
