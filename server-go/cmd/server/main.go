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
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"skillpass-server-go/internal/analytics"
	"skillpass-server-go/internal/application"
	"skillpass-server-go/internal/authtoken"
	"skillpass-server-go/internal/config"
	"skillpass-server-go/internal/db"
	_ "skillpass-server-go/docs"
	"skillpass-server-go/internal/email"
	"skillpass-server-go/internal/evaluation"
	"skillpass-server-go/internal/handlers"
	"skillpass-server-go/internal/hris/employee"
	"skillpass-server-go/internal/hris/org"
	"skillpass-server-go/internal/spdid"
	"skillpass-server-go/internal/lib"
	"skillpass-server-go/internal/matching"
	"skillpass-server-go/internal/middleware"
	"skillpass-server-go/internal/notification"
	"skillpass-server-go/internal/rbac"
	"skillpass-server-go/internal/resume"
	"skillpass-server-go/internal/storage"
	"skillpass-server-go/internal/webhook"

	"skillpass-server-go/internal/career"
	"skillpass-server-go/internal/companyreviews"
	"skillpass-server-go/internal/feedback"
	"skillpass-server-go/internal/profileviews"
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

	// Phase 4: email delivery + auth tokens + file storage
	emailSender := email.NewSender()
	tokenService := authtoken.NewService(database)
	auth.SetEmailer(emailSender)
	auth.SetTokenService(tokenService)

	store := storage.NewStore()
	uploads := handlers.NewUploadHandler(database, store)
	if ls, ok := store.(*storage.LocalStore); ok {
		r.Static("/uploads", ls.Dir())
	}

	// Phase 2: AI Evaluation & Matching
	llmClient := lib.NewLLMClient()
	evalService := evaluation.NewService(database, llmClient)
	evalHandler := evaluation.NewHandler(database, evalService)

	resumeService := resume.NewService(llmClient)
	resumeHandler := resume.NewHandler(resumeService)

	appService := application.NewService(database)
	appHandler := application.NewHandler(appService)

	notifService := notification.NewService(database)
	notifService.SetEmailer(emailSender)
	notifHandler := notification.NewHandler(notifService)
	appHandler.SetNotifier(notifService)

	webhookService := webhook.NewService(database)
	webhookHandler := webhook.NewHandler(webhookService)
	appHandler.SetWebhookDispatcher(webhookService)

	analyticsService := analytics.NewService(database)
	analyticsHandler := analytics.NewHandler(database, analyticsService)

	matchService := matching.NewService(database)
	matchHandler := matching.NewHandler(matchService)

	// Phase 3: Feedback & Career Growth
	feedbackService := feedback.NewService(database, llmClient)
	feedbackHandler := feedback.NewHandler(database, feedbackService)

	companyReviewsService := companyreviews.NewService(database)
	companyReviewsHandler := companyreviews.NewHandler(companyReviewsService)

	careerService := career.NewService(database, llmClient)
	careerHandler := career.NewHandler(database, careerService)

	profileViewsService := profileviews.NewService(database)
	profileViewsHandler := profileviews.NewHandler(database, profileViewsService)

	api.GET("/health", handlers.GetHealth)

	api.GET("/industries", ref.GetIndustries)
	api.GET("/tags", ref.GetTags)

	api.GET("/jobs", jobs.ListJobs)
	api.GET("/jobs/:id", jobs.GetJob)

	api.POST("/auth/register", authRL.Middleware(), auth.Register)
	api.POST("/auth/login", authRL.Middleware(), auth.Login)
	api.POST("/auth/refresh", authRL.Middleware(), auth.Refresh)
	api.POST("/auth/logout", middleware.AuthRequired(cfg.JWTSecret), auth.Logout)
	api.GET("/auth/me", middleware.AuthRequired(cfg.JWTSecret), auth.Me)
	api.GET("/auth/verify-email", auth.VerifyEmail)
	api.POST("/auth/resend-verification", middleware.AuthRequired(cfg.JWTSecret), auth.ResendVerification)
	api.POST("/auth/forgot-password", authRL.Middleware(), auth.ForgotPassword)
	api.POST("/auth/reset-password", authRL.Middleware(), auth.ResetPassword)

	api.GET("/profiles/:username", passport.GetProfile)

	// Server-rendered Open Graph page for link crawlers (outside /api/v1).
	r.GET("/p/:username", passport.GetOGPage)

	authGroup := api.Group("/profiles")
	authGroup.Use(middleware.AuthRequired(cfg.JWTSecret))
	authGroup.GET("/me", profiles.GetMyProfile)
	authGroup.PUT("/me", profiles.UpdateMyProfile)
	authGroup.POST("/me/experience", profiles.CreateExperience)
	authGroup.PUT("/me/experience/:id", profiles.UpdateExperience)
	authGroup.DELETE("/me/experience/:id", profiles.DeleteExperience)
	authGroup.POST("/me/resume-parse", resumeHandler.ParseResume)
	authGroup.POST("/me/resume-upload", resumeHandler.UploadResume)
	authGroup.POST("/me/avatar", uploads.UploadAvatar)
	authGroup.GET("/me/analytics", analyticsHandler.JobseekerAnalytics)

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
	evalGroup.POST("/me/career-path", evalHandler.PostCareerPath)

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
	companyAppGroup.GET("/analytics", analyticsHandler.CompanyAnalytics)
	companyAppGroup.GET("/webhooks", webhookHandler.List)
	companyAppGroup.POST("/webhooks", webhookHandler.Create)
	companyAppGroup.DELETE("/webhooks/:id", webhookHandler.Delete)

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
	matchesJobseekerGroup.GET("/:id/skills-gap", matchHandler.SkillsGap)

	matchesCompanyGroup := api.Group("/candidates")
	for _, m := range verifiedCompany {
		matchesCompanyGroup.Use(m)
	}
	matchesCompanyGroup.GET("/matches", matchHandler.MatchCandidates)

	// ── Phase 3: Feedback & Career Growth ──

	// Feedback routes (company gives feedback)
	feedbackCompanyGroup := api.Group("/feedback")
	for _, m := range verifiedCompany {
		feedbackCompanyGroup.Use(m)
	}
	feedbackCompanyGroup.GET("", feedbackHandler.GetCompanyFeedback)
	feedbackCompanyGroup.POST("/:profile_id", feedbackHandler.PostFeedback)

	// Feedback routes (jobseeker views feedback)
	feedbackJobseekerGroup := api.Group("/feedback")
	feedbackJobseekerGroup.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RequireRole("jobseeker"))
	feedbackJobseekerGroup.GET("/me", feedbackHandler.GetMyFeedback)
	feedbackJobseekerGroup.GET("/suggestions/me", feedbackHandler.GetMySuggestions)

	// Company reviews (candidate rates company)
	reviewsGroup := api.Group("/companies")
	reviewsGroup.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RequireRole("jobseeker"))
	reviewsGroup.POST("/:id/reviews", companyReviewsHandler.PostReview)
	reviewsGroup.GET("/:id/reviews", companyReviewsHandler.ListReviews)

	// Company reputation (public)
	api.GET("/companies/:id/reputation", companyReviewsHandler.GetReputation)

	// Career paths (authenticated)
	careerGroup := api.Group("/career")
	careerGroup.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RequireRole("jobseeker"))
	careerGroup.GET("/paths", careerHandler.ListCareerPaths)
	careerGroup.GET("/skill-gap/me", careerHandler.GetSkillGap)
	careerGroup.GET("/path/me", careerHandler.GetCareerPath)

	// Profile views (jobseeker)
	profileViewsGroup := api.Group("/profiles")
	profileViewsGroup.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RequireRole("jobseeker"))
	profileViewsGroup.GET("/me/views", profileViewsHandler.GetMyProfileViews)

	// Record profile views (company views profile)
	api.POST("/profiles/:profile_id/view", middleware.AuthRequired(cfg.JWTSecret), middleware.RequireRole("company"), profileViewsHandler.RecordView)

	// ── HRIS routes ──
	rbacService := rbac.NewService(database)
	empHandler := employee.NewHandler(database)
	orgHandler := org.NewHandler(database)

	hris := api.Group("/hris")
	hris.Use(middleware.AuthRequired(cfg.JWTSecret), rbac.RequireCompanyMember(rbacService))

	hrisEmployees := hris.Group("/employees")
	hrisEmployees.GET("", rbac.RequirePermission(rbacService, "employee.view", "employee.view_team"), empHandler.List)
	hrisEmployees.POST("", rbac.RequirePermission(rbacService, "employee.create"), empHandler.Create)
	hrisEmployees.GET("/:id", rbac.RequirePermission(rbacService, "employee.view", "employee.view_team", "employee.view_self"), empHandler.Get)
	hrisEmployees.PUT("/:id", rbac.RequirePermission(rbacService, "employee.update"), empHandler.Update)

	hrisBranches := hris.Group("/branches")
	hrisBranches.GET("", rbac.RequirePermission(rbacService, "org.view"), orgHandler.ListBranches)
	hrisBranches.POST("", rbac.RequirePermission(rbacService, "org.manage"), orgHandler.CreateBranch)
	hrisBranches.GET("/:id", rbac.RequirePermission(rbacService, "org.view"), orgHandler.GetBranch)
	hrisBranches.PUT("/:id", rbac.RequirePermission(rbacService, "org.manage"), orgHandler.UpdateBranch)
	hrisBranches.DELETE("/:id", rbac.RequirePermission(rbacService, "org.manage"), orgHandler.DeleteBranch)

	hrisDepts := hris.Group("/departments")
	hrisDepts.GET("", rbac.RequirePermission(rbacService, "org.view"), orgHandler.ListDepartments)
	hrisDepts.POST("", rbac.RequirePermission(rbacService, "org.manage"), orgHandler.CreateDepartment)
	hrisDepts.PUT("/:id", rbac.RequirePermission(rbacService, "org.manage"), orgHandler.UpdateDepartment)
	hrisDepts.DELETE("/:id", rbac.RequirePermission(rbacService, "org.manage"), orgHandler.DeleteDepartment)

	hrisPositions := hris.Group("/positions")
	hrisPositions.GET("", rbac.RequirePermission(rbacService, "org.view"), orgHandler.ListPositions)
	hrisPositions.POST("", rbac.RequirePermission(rbacService, "org.manage"), orgHandler.CreatePosition)
	hrisPositions.PUT("/:id", rbac.RequirePermission(rbacService, "org.manage"), orgHandler.UpdatePosition)
	hrisPositions.DELETE("/:id", rbac.RequirePermission(rbacService, "org.manage"), orgHandler.DeletePosition)

	hris.GET("/org/tree", rbac.RequirePermission(rbacService, "org.view"), orgHandler.GetOrgTree)
	hris.GET("/org/chart", rbac.RequirePermission(rbacService, "org.view"), orgHandler.GetOrgChart)

	// SP-DID
	spdidHandler := spdid.NewHandler(database)
	hrisEmployees.POST("/:id/did", rbac.RequirePermission(rbacService, "org.manage"), spdidHandler.CreateDID)
	hrisEmployees.GET("/:id/did", rbac.RequirePermission(rbacService, "employee.view"), spdidHandler.GetDID)

	// Working Calendars
	hrisCalendars := hris.Group("/working-calendars")
	hrisCalendars.POST("", rbac.RequirePermission(rbacService, "org.manage"), orgHandler.CreateCalendar)
	hrisCalendars.GET("", rbac.RequirePermission(rbacService, "org.view"), orgHandler.ListCalendars)
	hrisCalendars.PUT("/:id", rbac.RequirePermission(rbacService, "org.manage"), orgHandler.UpdateCalendar)
	hrisCalendars.DELETE("/:id", rbac.RequirePermission(rbacService, "org.manage"), orgHandler.DeleteCalendar)

	hrisRoles := hris.Group("/roles")
	hrisRoles.GET("", rbac.RequirePermission(rbacService, "org.view"), func(c *gin.Context) {
		cid, err := uuid.Parse(c.GetString("companyId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
			return
		}
		roles, err := rbacService.ListRoles(c.Request.Context(), cid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list roles"})
			return
		}
		c.JSON(http.StatusOK, roles)
	})

	hrisEmployeeRoles := hris.Group("/employees/:id/roles")
	hrisEmployeeRoles.POST("", rbac.RequirePermission(rbacService, "roles.manage"), func(c *gin.Context) {
		companyID, err := uuid.Parse(c.GetString("companyId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
			return
		}
		employeeID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
			return
		}
		var req struct {
			RoleID uuid.UUID `json:"roleId" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := rbacService.AssignRole(c.Request.Context(), companyID, employeeID, req.RoleID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign role"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Role assigned"})
	})
	hrisEmployeeRoles.DELETE("/:roleId", rbac.RequirePermission(rbacService, "roles.manage"), func(c *gin.Context) {
		companyID, err := uuid.Parse(c.GetString("companyId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
			return
		}
		employeeID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
			return
		}
		roleID, err := uuid.Parse(c.Param("roleId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
			return
		}
		if err := rbacService.RemoveRole(c.Request.Context(), companyID, employeeID, roleID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove role"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Role removed"})
	})

	hris.GET("/me/permissions", func(c *gin.Context) {
		employeeID, err := uuid.Parse(c.GetString("employeeId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
			return
		}
		perms, err := rbacService.GetEmployeePermissions(c.Request.Context(), employeeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get permissions"})
			return
		}
		roles, err := rbacService.GetEmployeeRoles(c.Request.Context(), employeeID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get roles"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"permissions": perms, "roles": roles})
	})

	// Swagger UI
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Serve static files (production: embedded SPA, dev: no-op)
	if cfg.ServeStatic {
		setupStatic(r)
	}

	log.Printf("SkillPass API running at http://localhost:%s", cfg.Port)
	if cfg.ServeStatic {
		log.Printf("Frontend served from embedded files at http://localhost:%s", cfg.Port)
	} else {
		log.Printf("Frontend: use Vite dev server (bun run dev:web)")
	}
	log.Printf("Swagger UI at http://localhost:%s/docs/index.html", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
