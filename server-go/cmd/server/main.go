// SkillPass API
//
// SkillPass — skills-based hiring platform API.
//
//	@title			SkillPass API
//	@version		1.0
//	@description	SkillPass — skills-based hiring platform API
//	@host			localhost:1234
//	@BasePath		/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
package main

import (
	"context"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"skillpass-server-go/internal/application"
	"skillpass-server-go/internal/config"
	"skillpass-server-go/internal/db"
	_ "skillpass-server-go/docs"
	"skillpass-server-go/internal/evaluation"
	"skillpass-server-go/internal/handlers"
	"skillpass-server-go/internal/lib"
	"skillpass-server-go/internal/matching"
	"skillpass-server-go/internal/middleware"
	"skillpass-server-go/internal/notification"
	"skillpass-server-go/internal/resume"
)

func main() {
	_ = godotenv.Load(".env", "../.env")

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
		ExposeHeaders:    []string{"Set-Cookie"},
	}))

	authRL := middleware.NewRateLimiter(5, 10)

	api := r.Group("/api/v1")

	ref := handlers.NewReferenceHandler(database)
	jobs := handlers.NewJobHandler(database)
	auth := handlers.NewAuthHandler(database, cfg.JWTSecret)
	profiles := handlers.NewProfileHandler(database)
	passport := handlers.NewPassportHandler(database)
	companies := handlers.NewCompanyHandler(database)
	search := handlers.NewSearchHandler(database)
	admin := handlers.NewAdminHandler(database)

	// Phase 2: AI Evaluation & Matching
	llmClient := lib.NewLLMClient()
	evalService := evaluation.NewService(database, llmClient)
	evalHandler := evaluation.NewHandler(database, evalService)

	resumeService := resume.NewService(llmClient)
	resumeHandler := resume.NewHandler(resumeService)

	appService := application.NewService(database)
	appHandler := application.NewHandler(appService)

	notifService := notification.NewService(database)
	notifHandler := notification.NewHandler(notifService)
	appHandler.SetNotifier(notifService)

	matchService := matching.NewService(database)
	matchHandler := matching.NewHandler(matchService)

	api.GET("/health", handlers.GetHealth)

	api.GET("/industries", ref.GetIndustries)
	api.GET("/tags", ref.GetTags)

	api.GET("/jobs", jobs.ListJobs)
	api.GET("/jobs/:id", jobs.GetJob)

	api.POST("/auth/register", authRL.Middleware(), auth.Register)
	api.POST("/auth/login", authRL.Middleware(), auth.Login)
	api.POST("/auth/refresh", authRL.Middleware(), auth.Refresh)
	api.POST("/auth/logout", middleware.AuthRequired(cfg.JWTSecret), auth.Logout)

	api.GET("/profiles/:username", passport.GetProfile)

	authGroup := api.Group("/profiles")
	authGroup.Use(middleware.AuthRequired(cfg.JWTSecret))
	authGroup.GET("/me", profiles.GetMyProfile)
	authGroup.PUT("/me", profiles.UpdateMyProfile)
	authGroup.POST("/me/experience", profiles.CreateExperience)
	authGroup.PUT("/me/experience/:id", profiles.UpdateExperience)
	authGroup.DELETE("/me/experience/:id", profiles.DeleteExperience)
	authGroup.POST("/me/resume-parse", resumeHandler.ParseResume)

	companyGroup := api.Group("/company")
	companyGroup.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RequireRole("company"))
	companyGroup.GET("/profile", companies.GetProfile)
	companyGroup.PUT("/profile", companies.UpdateProfile)
	companyGroup.POST("/verification", companies.SubmitVerification)
	companyGroup.GET("/verification-status", companies.GetVerificationStatus)

	verifiedCompany := []gin.HandlerFunc{
		middleware.AuthRequired(cfg.JWTSecret),
		middleware.RequireRole("company"),
		middleware.RequireVerifiedCompany(database),
	}

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

	adminGroup := api.Group("/admin")
	adminGroup.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RequireRole("admin"))
	adminGroup.GET("/verifications/pending", admin.ListPendingVerifications)
	adminGroup.POST("/verifications/:id", admin.HandleVerification)

	// ── Evaluation routes (jobseeker) ──
	evalGroup := api.Group("/evaluate")
	evalGroup.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RequireRole("jobseeker"))
	evalGroup.POST("/me", evalHandler.PostEvaluate)
	evalGroup.GET("/me/results", evalHandler.GetLatestEvaluation)

	// ── Application routes (jobseeker applies) ──
	jobApplyGroup := api.Group("/jobs")
	jobApplyGroup.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RequireRole("jobseeker"))
	jobApplyGroup.POST("/:id/apply", appHandler.Apply)

	appGroup := api.Group("/applications")
	appGroup.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RequireRole("jobseeker"))
	appGroup.GET("/me", appHandler.ListMyApplications)

	// ── Company application management (verified company) ──
	companyAppGroup := api.Group("/company")
	for _, m := range verifiedCompany {
		companyAppGroup.Use(m)
	}
	companyAppGroup.GET("/applications", appHandler.ListCompanyApplications)

	appStatusGroup := api.Group("/applications")
	for _, m := range verifiedCompany {
		appStatusGroup.Use(m)
	}
	appStatusGroup.PUT("/:id/status", appHandler.UpdateStatus)
	appStatusGroup.GET("/:id/messages", appHandler.ListMessages)
	appStatusGroup.POST("/:id/messages", appHandler.AddMessage)

	// ── Notification routes (any authenticated user) ──
	notifGroup := api.Group("/notifications")
	notifGroup.Use(middleware.AuthRequired(cfg.JWTSecret))
	notifGroup.GET("/me", notifHandler.ListMine)
	notifGroup.PUT("/read-all", notifHandler.MarkAllRead)
	notifGroup.PUT("/:id/read", notifHandler.MarkRead)

	// ── Matching routes ──
	matchesJobseekerGroup := api.Group("/jobs")
	matchesJobseekerGroup.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RequireRole("jobseeker"))
	matchesJobseekerGroup.GET("/matches", matchHandler.MatchJobs)

	matchesCompanyGroup := api.Group("/candidates")
	for _, m := range verifiedCompany {
		matchesCompanyGroup.Use(m)
	}
	matchesCompanyGroup.GET("/matches", matchHandler.MatchCandidates)

	// Swagger UI
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Printf("SkillPass API (Go) running at http://localhost:%s", cfg.Port)
	log.Printf("Swagger UI at http://localhost:%s/docs/index.html", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
