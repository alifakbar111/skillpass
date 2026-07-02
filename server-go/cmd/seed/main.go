package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	. "github.com/go-jet/jet/v2/postgres"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	"skillpass-server-go/.gen/skillpass/public/model"
	"skillpass-server-go/internal/gen"
)

func main() {
	_ = godotenv.Load(".env", "../.env")

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/skillpass"
	}

	ctx := context.Background()

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping DB: %v", err)
	}

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

	count := 0
	for _, ind := range industries {
		stmt := gen.IndustryCategories.INSERT(
			gen.IndustryCategories.Name, gen.IndustryCategories.Description,
		).VALUES(ind.Name, ind.Description).
			ON_CONFLICT(gen.IndustryCategories.Name).DO_NOTHING()

		_, err := stmt.ExecContext(ctx, db)
		if err != nil {
			log.Printf("  Warning: failed to insert industry %s: %v", ind.Name, err)
			continue
		}
		count++
	}
	fmt.Printf("Seeded %d industry categories\n", count)

	// Seed skills across multiple industries
	skillNames := []string{
		// Technology
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
		// Healthcare
		"Patient Care", "Medical Records", "HIPAA Compliance", "Clinical Research",
		"EHR Systems", "Medical Coding", "Phlebotomy", "Vital Signs Monitoring",
		"Infection Control", "Radiology", "Pharmacology", "Patient Assessment",
		"Care Coordination", "ICD-10", "Telemedicine", "EMR Systems",
		// Finance & Accounting
		"Financial Analysis", "Budgeting", "Forecasting", "QuickBooks",
		"Tax Preparation", "Auditing", "Risk Management", "GAAP",
		"Financial Reporting", "Payroll", "Internal Controls", "ERP Systems",
		"Accounts Payable", "Accounts Receivable", "Reconciliation", "SAP",
		// Marketing & Sales
		"SEO", "Content Marketing", "Social Media", "Google Analytics",
		"CRM", "Sales Strategy", "Lead Generation", "Email Marketing",
		"PPC Advertising", "Brand Management", "Market Research", "Copywriting",
		"HubSpot", "Salesforce", "Public Relations", "Digital Marketing",
		// Manufacturing & Engineering
		"Lean Manufacturing", "Six Sigma", "CAD", "SolidWorks",
		"Supply Chain Management", "PLC Programming", "Quality Assurance",
		"Process Improvement", "CNC Operation", "OSHA Compliance",
		"Inventory Management", "AutoCAD", "Root Cause Analysis", "Kaizen",
		// Education
		"Curriculum Development", "Classroom Management", "Lesson Planning",
		"Student Assessment", "Educational Technology", "Special Education",
		"ESL Instruction", "Online Teaching", "Learning Management Systems",
		"Academic Advising", "Early Childhood Education", "Grant Writing",
		// Legal
		"Legal Research", "Contract Review", "Case Management", "Litigation Support",
		"Legal Writing", "Discovery", "Compliance", "Intellectual Property",
		"Corporate Law", "Due Diligence", "Regulatory Affairs", "Mediation",
		// Hospitality
		"Customer Service", "Event Planning", "Food Safety", "Front Desk Operations",
		"Housekeeping Management", "Reservation Systems", "Banquet Management",
		"Menu Planning", "POS Systems", "Concierge Services", "Travel Coordination",
		// General Business
		"Strategic Planning", "Data Analysis", "Communication", "Negotiation",
		"Problem Solving", "Team Building", "Time Management",
		"Microsoft Excel", "Microsoft PowerPoint", "Microsoft Word", "Public Speaking",
		"Business Development", "Operations Management", "Vendor Management",
	}

	skillCount := 0
	for _, name := range skillNames {
		stmt := gen.Skills.INSERT(gen.Skills.Name).VALUES(name).
			ON_CONFLICT(gen.Skills.Name).DO_NOTHING()
		_, err := stmt.ExecContext(ctx, db)
		if err != nil {
			log.Printf("  Warning: failed to insert skill %s: %v", name, err)
			continue
		}
		skillCount++
	}
	fmt.Printf("Seeded %d skills\n", skillCount)

	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")

	if adminEmail == "" && adminPassword == "" {
		fmt.Println("Skipping admin seed (ADMIN_EMAIL and ADMIN_PASSWORD not set)")
		return
	}
	if adminEmail == "" || adminPassword == "" {
		log.Fatal("Both ADMIN_EMAIL and ADMIN_PASSWORD must be set to seed an admin user")
	}

	checkStmt := SELECT(
		gen.Users.ID,
	).FROM(
		gen.Users,
	).WHERE(
		gen.Users.Email.EQ(String(adminEmail)),
	)

	var existing model.Users
	err = checkStmt.QueryContext(ctx, db, &existing)
	if err == nil {
		fmt.Println("Admin user already exists, skipping")
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	insertStmt := gen.Users.INSERT(
		gen.Users.Email, gen.Users.Username, gen.Users.PasswordHash,
		gen.Users.Name, gen.Users.Role,
	).VALUES(
		adminEmail, "admin", string(passwordHash), "Admin", "admin",
	)

	_, err = insertStmt.ExecContext(ctx, db)
	if err != nil {
		log.Fatalf("Failed to create admin: %v", err)
	}
	fmt.Printf("Seeded admin user (%s)\n", adminEmail)
}
