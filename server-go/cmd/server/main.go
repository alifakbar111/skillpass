package main

import (
	"context"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"skillpass-server-go/internal/config"
	"skillpass-server-go/internal/db"
	"skillpass-server-go/internal/handlers"
	"skillpass-server-go/internal/middleware"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	ctx := context.Background()
	database, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.CORSOrigin},
		AllowCredentials: true,
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	}))

	api := r.Group("/api/v1")

	// ── Handlers ──
	ref := handlers.NewReferenceHandler(database)
	jobs := handlers.NewJobHandler(database)
	auth := handlers.NewAuthHandler(database, cfg.JWTSecret)
	profiles := handlers.NewProfileHandler(database)
	passport := handlers.NewPassportHandler(database)
	companies := handlers.NewCompanyHandler(database)
	search := handlers.NewSearchHandler(database)
	admin := handlers.NewAdminHandler(database)

	// ── Public routes ──
	api.GET("/health", handlers.GetHealth)

	api.GET("/industries", ref.GetIndustries)
	api.GET("/tags", ref.GetTags)

	api.GET("/jobs", jobs.ListJobs)
	api.GET("/jobs/:id", jobs.GetJob)

	api.POST("/auth/register", auth.Register)
	api.POST("/auth/login", auth.Login)
	api.POST("/auth/refresh", auth.Refresh)
	api.POST("/auth/logout", auth.Logout)

	api.GET("/profiles/:username", passport.GetProfile)

	// ── Authenticated (any role) ──
	authGroup := api.Group("/profiles")
	authGroup.Use(middleware.AuthRequired(cfg.JWTSecret))
	authGroup.GET("/me", profiles.GetMyProfile)
	authGroup.PUT("/me", profiles.UpdateMyProfile)
	authGroup.POST("/me/experience", profiles.CreateExperience)
	authGroup.PUT("/me/experience/:id", profiles.UpdateExperience)
	authGroup.DELETE("/me/experience/:id", profiles.DeleteExperience)

	// ── Company only ──
	companyGroup := api.Group("/company")
	companyGroup.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RequireRole("company"))
	companyGroup.GET("/profile", companies.GetProfile)
	companyGroup.PUT("/profile", companies.UpdateProfile)
	companyGroup.POST("/verification", companies.SubmitVerification)
	companyGroup.GET("/verification-status", companies.GetVerificationStatus)

	// ── Verified company only ──
	verifiedCompany := append([]gin.HandlerFunc{},
		middleware.AuthRequired(cfg.JWTSecret),
		middleware.RequireRole("company"),
		middleware.RequireVerifiedCompany(database),
	)

	jobsGroup := api.Group("/jobs")
	for _, m := range verifiedCompany {
		jobsGroup.Use(m)
	}
	jobsGroup.GET("/me", jobs.ListMyJobs)
	jobsGroup.POST("", jobs.CreateJob)
	jobsGroup.PUT("/:id", jobs.UpdateJob)
	jobsGroup.DELETE("/:id", jobs.DeleteJob)

	searchGroup := api.Group("/search")
	for _, m := range verifiedCompany {
		searchGroup.Use(m)
	}
	searchGroup.GET("/candidates", search.SearchCandidates)

	// ── Admin only ──
	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RequireRole("admin"))
	adminGroup.GET("/verifications/pending", admin.ListPendingVerifications)
	adminGroup.POST("/verifications/:id", admin.HandleVerification)

	log.Printf("SkillPass API (Go) running at http://localhost:%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
