package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"golang.org/x/crypto/bcrypt"

	"skillpass-server-go/internal/models"
)

func main() {
	_ = godotenv.Load(".env", "../.env")

	databaseURL := os.Getenv("DATABASE_URL")

	ctx := context.Background()

	sqlDB, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer sqlDB.Close()

	if err := sqlDB.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}

	bunDB := bun.NewDB(sqlDB, pgdialect.New())
	defer bunDB.Close()

	fmt.Println("Seeding database...")

	industries := []struct {
		Name        string
		Description string
	}{
		{"Technology", "Software, hardware, IT services"},
		{"Manufacturing", "Industrial production and fabrication"},
		{"Healthcare", "Medical services and pharmaceuticals"},
		{"Finance", "Banking, investment, insurance"},
		{"Education", "Schools, universities, training"},
		{"Retail", "Sales, e-commerce, consumer goods"},
		{"Transportation", "Logistics, delivery, ride-hailing"},
		{"Creative Arts", "Design, media, entertainment"},
		{"Hospitality", "Hotels, restaurants, tourism"},
		{"Construction", "Building and infrastructure"},
		{"Agriculture", "Farming, food production"},
		{"Energy", "Oil, gas, renewable energy"},
	}

	industryModels := make([]models.IndustryCategory, len(industries))
	for i, ind := range industries {
		desc := ind.Description
		industryModels[i] = models.IndustryCategory{Name: ind.Name, Description: &desc}
	}
	res, err := bunDB.NewInsert().Model(&industryModels).On("CONFLICT (name) DO NOTHING").Exec(ctx)
	if err != nil {
		log.Printf("  Warning: failed to insert industries: %v", err)
	} else {
		n, _ := res.RowsAffected()
		fmt.Printf("Seeded %d industry categories\n", n)
	}

	skillNames := []string{
		"Go", "TypeScript", "JavaScript", "Python", "Java", "Rust",
		"React", "Vue.js", "Angular", "Next.js", "Node.js", "Express",
		"PostgreSQL", "MySQL", "MongoDB", "Redis", "SQLite",
		"Docker", "Kubernetes", "AWS", "GCP", "Azure", "Terraform",
		"Git", "CI/CD", "GraphQL", "REST API", "gRPC",
		"Tailwind CSS", "DaisyUI", "SASS", "CSS", "HTML",
		"Gin", "Echo", "FastAPI", "Django", "Flask",
		"React Native", "Flutter", "Swift", "Kotlin",
		"Machine Learning", "Data Science", "SQL", "NoSQL",
		"Linux", "Bash", "Nginx", "Apache", "RabbitMQ", "Kafka",
		"OAuth", "JWT", "SAML", "OpenID Connect", "RBAC",
		"Microservices", "Event-Driven Architecture", "DDD", "CQRS",
		"Elasticsearch", "Prometheus", "Grafana", "Datadog",
		"WebSockets", "Server-Sent Events", "WebRTC",
		"Accessibility", "WCAG", "a11y", "Performance Optimization",
		"Figma", "Sketch", "Adobe XD", "UI Design", "UX Design",
		"Agile", "Scrum", "Project Management", "Leadership",
		"Testing", "TDD", "Cypress", "Vitest", "Jest", "Playwright",
		"Patient Care", "Medical Records", "HIPAA Compliance", "Clinical Research",
		"EHR Systems", "Medical Coding", "Phlebotomy", "Vital Signs Monitoring",
		"Infection Control", "Radiology", "Pharmacology", "Patient Assessment",
		"Care Coordination", "ICD-10", "Telemedicine", "EMR Systems",
		"Financial Analysis", "Budgeting", "Forecasting", "QuickBooks",
		"Tax Preparation", "Auditing", "Risk Management", "GAAP",
		"Financial Reporting", "Payroll", "Internal Controls", "ERP Systems",
		"Accounts Payable", "Accounts Receivable", "Reconciliation", "SAP",
		"SEO", "Content Marketing", "Social Media", "Google Analytics",
		"CRM", "Sales Strategy", "Lead Generation", "Email Marketing",
		"PPC Advertising", "Brand Management", "Market Research", "Copywriting",
		"HubSpot", "Salesforce", "Public Relations", "Digital Marketing",
		"Lean Manufacturing", "Six Sigma", "CAD", "SolidWorks",
		"Supply Chain Management", "PLC Programming", "Quality Assurance",
		"Process Improvement", "CNC Operation", "OSHA Compliance",
		"Inventory Management", "AutoCAD", "Root Cause Analysis", "Kaizen",
		"Curriculum Development", "Classroom Management", "Lesson Planning",
		"Student Assessment", "Educational Technology", "Special Education",
		"ESL Instruction", "Online Teaching", "Learning Management Systems",
		"Academic Advising", "Early Childhood Education", "Grant Writing",
		"Legal Research", "Contract Review", "Case Management", "Litigation Support",
		"Legal Writing", "Discovery", "Compliance", "Intellectual Property",
		"Corporate Law", "Due Diligence", "Regulatory Affairs", "Mediation",
		"Customer Service", "Event Planning", "Food Safety", "Front Desk Operations",
		"Housekeeping Management", "Reservation Systems", "Banquet Management",
		"Menu Planning", "POS Systems", "Concierge Services", "Travel Coordination",
		"Strategic Planning", "Data Analysis", "Communication", "Negotiation",
		"Problem Solving", "Team Building", "Time Management",
		"Microsoft Excel", "Microsoft PowerPoint", "Microsoft Word", "Public Speaking",
		"Business Development", "Operations Management", "Vendor Management",
	}

	now := time.Now()
	skillModels := make([]models.Skill, len(skillNames))
	for i, name := range skillNames {
		skillModels[i] = models.Skill{Name: name, CreatedAt: now}
	}
	res, err = bunDB.NewInsert().Model(&skillModels).On("CONFLICT (name) DO NOTHING").Exec(ctx)
	if err != nil {
		log.Printf("  Warning: failed to insert skills: %v", err)
	} else {
		n, _ := res.RowsAffected()
		fmt.Printf("Seeded %d skills\n", n)
	}

	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")

	if adminEmail == "" && adminPassword == "" {
		fmt.Println("Skipping admin seed (ADMIN_EMAIL and ADMIN_PASSWORD not set)")
		return
	}
	if adminEmail == "" || adminPassword == "" {
		log.Fatal("Both ADMIN_EMAIL and ADMIN_PASSWORD must be set to seed an admin user")
	}

	existing := new(models.User)
	err = bunDB.NewSelect().Model(existing).Where("email = ?", adminEmail).Scan(ctx)
	if err == nil {
		fmt.Println("Admin user already exists, skipping")
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	admin := &models.User{
		Email:        adminEmail,
		Username:     "admin",
		PasswordHash: string(passwordHash),
		Name:         "Admin",
		Role:         "admin",
		CreatedAt:    time.Now(),
	}
	_, err = bunDB.NewInsert().Model(admin).Exec(ctx)
	if err != nil {
		log.Fatalf("Failed to create admin: %v", err)
	}
	fmt.Printf("Seeded admin user (%s)\n", adminEmail)
}
