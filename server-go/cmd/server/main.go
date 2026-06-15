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
	"skillpass-server-go/internal/hris/attendance"
	"skillpass-server-go/internal/hris/employee"
	"skillpass-server-go/internal/hris/holiday"
	"skillpass-server-go/internal/hris/leave"
	"skillpass-server-go/internal/hris/org"
	"skillpass-server-go/internal/hris/payroll"
	"skillpass-server-go/internal/hris/shift"
	"skillpass-server-go/internal/spdid"
	"skillpass-server-go/internal/lib"
	"skillpass-server-go/internal/matching"
	"skillpass-server-go/internal/middleware"
	"skillpass-server-go/internal/notification"
	"skillpass-server-go/internal/rbac"
	"skillpass-server-go/internal/resume"
	"skillpass-server-go/internal/storage"
	"skillpass-server-go/internal/webhook"
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

	// Shift Templates
	shiftHandler := shift.NewHandler(database)
	hrisShifts := hris.Group("/shifts")
	hrisShifts.GET("", rbac.RequirePermission(rbacService, "org.view"), shiftHandler.ListTemplates)
	hrisShifts.POST("", rbac.RequirePermission(rbacService, "org.manage"), shiftHandler.CreateTemplate)
	hrisShifts.PUT("/:id", rbac.RequirePermission(rbacService, "org.manage"), shiftHandler.UpdateTemplate)
	hrisShifts.DELETE("/:id", rbac.RequirePermission(rbacService, "org.manage"), shiftHandler.DeleteTemplate)
	hrisEmployees.POST("/:id/shifts", rbac.RequirePermission(rbacService, "employee.update"), shiftHandler.AssignShift)
	hrisEmployees.GET("/:id/shifts", rbac.RequirePermission(rbacService, "employee.view"), shiftHandler.ListEmployeeShifts)

	// Attendance
	attHandler := attendance.NewHandler(database)
	hrisAttendance := hris.Group("/attendance")
	hrisAttendance.POST("/clock-in", rbac.RequirePermission(rbacService, "attendance.clock"), attHandler.ClockIn)
	hrisAttendance.POST("/clock-out", rbac.RequirePermission(rbacService, "attendance.clock"), attHandler.ClockOut)
	hrisAttendance.GET("/dashboard", rbac.RequirePermission(rbacService, "attendance.view"), attHandler.Dashboard)
	hrisAttendance.GET("/my", rbac.RequirePermission(rbacService, "attendance.view_self"), attHandler.MyAttendance)
	hrisAttendance.GET("/ws", attHandler.Hub().HandleWS)

	// Attendance Exceptions
	hrisExceptions := hris.Group("/attendance-exceptions")
	hrisExceptions.POST("", rbac.RequirePermission(rbacService, "attendance.clock"), attHandler.CreateException)
	hrisExceptions.GET("", rbac.RequirePermission(rbacService, "attendance.view"), attHandler.ListExceptions)
	hrisExceptions.PUT("/:id/review", rbac.RequirePermission(rbacService, "attendance.manage"), attHandler.ReviewException)

	// Leave Types
	leaveHandler := leave.NewHandler(database)
	hrisLeaveTypes := hris.Group("/leave-types")
	hrisLeaveTypes.GET("", rbac.RequirePermission(rbacService, "org.view"), leaveHandler.ListTypes)
	hrisLeaveTypes.POST("", rbac.RequirePermission(rbacService, "org.manage"), leaveHandler.CreateType)
	hrisLeaveTypes.PUT("/:id", rbac.RequirePermission(rbacService, "org.manage"), leaveHandler.UpdateType)
	hrisLeaveTypes.DELETE("/:id", rbac.RequirePermission(rbacService, "org.manage"), leaveHandler.DeleteType)

	// Leave Balances
	hrisEmployees.GET("/:id/leave-balances", rbac.RequirePermission(rbacService, "employee.view", "employee.view_self"), leaveHandler.GetBalances)
	hrisEmployees.POST("/:id/leave-balances/init", rbac.RequirePermission(rbacService, "employee.update"), leaveHandler.InitBalances)

	// Leave Requests
	hrisLeave := hris.Group("/leave-requests")
	hrisLeave.POST("", rbac.RequirePermission(rbacService, "leave.request"), leaveHandler.CreateRequest)
	hrisLeave.GET("", rbac.RequirePermission(rbacService, "leave.view"), leaveHandler.ListRequests)
	hrisLeave.GET("/my", rbac.RequirePermission(rbacService, "leave.request"), leaveHandler.MyRequests)
	hrisLeave.PUT("/:id/review", rbac.RequirePermission(rbacService, "leave.manage"), leaveHandler.ReviewRequest)
	hrisLeave.PUT("/:id/cancel", rbac.RequirePermission(rbacService, "leave.request"), leaveHandler.CancelRequest)

	// Holidays
	holidayHandler := holiday.NewHandler(database)
	hrisHolidays := hris.Group("/holidays")
	hrisHolidays.GET("", rbac.RequirePermission(rbacService, "org.view"), holidayHandler.List)
	hrisHolidays.POST("", rbac.RequirePermission(rbacService, "org.manage"), holidayHandler.Create)
	hrisHolidays.PUT("/:id", rbac.RequirePermission(rbacService, "org.manage"), holidayHandler.Update)
	hrisHolidays.DELETE("/:id", rbac.RequirePermission(rbacService, "org.manage"), holidayHandler.Delete)

	// Payroll
	payrollHandler := payroll.NewHandler(database)
	hrisComponents := hris.Group("/salary-components")
	hrisComponents.GET("", rbac.RequirePermission(rbacService, "payroll.view", "payroll.manage"), payrollHandler.ListComponents)
	hrisComponents.POST("", rbac.RequirePermission(rbacService, "payroll.manage"), payrollHandler.CreateComponent)
	hrisComponents.PUT("/:id", rbac.RequirePermission(rbacService, "payroll.manage"), payrollHandler.UpdateComponent)
	hrisComponents.DELETE("/:id", rbac.RequirePermission(rbacService, "payroll.manage"), payrollHandler.DeleteComponent)

	hrisEmployees.GET("/:id/salary", rbac.RequirePermission(rbacService, "payroll.view", "payroll.manage"), payrollHandler.GetEmployeeSalary)
	hrisEmployees.PUT("/:id/salary", rbac.RequirePermission(rbacService, "payroll.manage"), payrollHandler.SetEmployeeSalary)

	hrisPayroll := hris.Group("/payroll-runs")
	hrisPayroll.GET("", rbac.RequirePermission(rbacService, "payroll.view", "payroll.run"), payrollHandler.ListRuns)
	hrisPayroll.POST("", rbac.RequirePermission(rbacService, "payroll.run"), payrollHandler.CreateRun)
	hrisPayroll.POST("/:id/calculate", rbac.RequirePermission(rbacService, "payroll.run"), payrollHandler.CalculateRun)
	hrisPayroll.POST("/:id/approve", rbac.RequirePermission(rbacService, "payroll.approve"), payrollHandler.ApproveRun)
	hrisPayroll.POST("/:id/mark-paid", rbac.RequirePermission(rbacService, "payroll.approve"), payrollHandler.MarkPaid)
	hrisPayroll.GET("/:id/payslips", rbac.RequirePermission(rbacService, "payroll.view"), payrollHandler.ListPayslips)

	hrisPayslips := hris.Group("/payslips")
	hrisPayslips.GET("/my", rbac.RequirePermission(rbacService, "payroll.view_self"), payrollHandler.MyPayslips)
	hrisPayslips.GET("/:payslipId", rbac.RequirePermission(rbacService, "payroll.view", "payroll.view_self"), payrollHandler.GetPayslip)

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
