package evaluation

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"skillpass-server-go/internal/lib"
)

// testProfile represents test profile input for prompt testing.
type testProfile struct {
	Name              string
	Headline          string
	About             string
	YearsOfExperience int
	Experiences       []testExperience
}

type testExperience struct {
	Type        string
	Title       string
	Organization string
	StartDate   string
	EndDate     string
	IsCurrent   bool
	Description string
	Industry    string
	SkillsUsed  []string
	URL         string
}

// buildTestUserPrompt simulates the user prompt the LLM receives.
func buildTestUserPrompt(p testProfile) string {
	var expLines []string
	for _, e := range p.Experiences {
		end := e.EndDate
		if end == "" && e.IsCurrent {
			end = "present"
		}
		skills := strings.Join(e.SkillsUsed, ", ")
		line := fmt.Sprintf("- %s at %s (%s to %s) [%s] Skills: %s",
			e.Title, e.Organization, e.StartDate, end, e.Type, skills)
		if e.Description != "" {
			line += "\n  Description: " + e.Description
		}
		if e.URL != "" {
			line += "\n  URL: " + e.URL
		}
		expLines = append(expLines, line)
	}
	allSkills := map[string]bool{}
	for _, e := range p.Experiences {
		for _, s := range e.SkillsUsed {
			allSkills[s] = true
		}
	}
	skillList := make([]string, 0, len(allSkills))
	for s := range allSkills {
		skillList = append(skillList, s)
	}
	return fmt.Sprintf(`Extract skill facts from this jobseeker profile:

Name: %s
Headline: %s
About: %s
Years of Experience: %d

Experience entries:
%s

Skills mentioned across experiences: %s

Return the extracted facts as JSON per the system prompt schema.`,
		p.Name, p.Headline, p.About, p.YearsOfExperience,
		strings.Join(expLines, "\n"),
		strings.Join(skillList, ", "))
}

func testTechProfile() testProfile {
	return testProfile{
		Name:              "Alice Chen",
		Headline:          "Senior Full-Stack Engineer",
		About:             "Experienced engineer with a focus on Go and React.",
		YearsOfExperience: 7,
		Experiences: []testExperience{
			{
				Type:         "employment",
				Title:        "Senior Backend Engineer",
				Organization: "TechCorp",
				StartDate:    "2021-03",
				EndDate:      "",
				IsCurrent:    true,
				Description:  "Lead the API team, designed microservices architecture, mentored 4 junior engineers.",
				Industry:     "Technology",
				SkillsUsed:   []string{"Go", "PostgreSQL", "Docker", "Kubernetes", "gRPC"},
				URL:          "https://linkedin.com/in/alice",
			},
			{
				Type:         "employment",
				Title:        "Backend Developer",
				Organization: "StartupXYZ",
				StartDate:    "2018-06",
				EndDate:      "2021-02",
				IsCurrent:    false,
				Description:  "Built REST APIs in Go, managed PostgreSQL schemas, set up CI/CD pipelines.",
				Industry:     "Technology",
				SkillsUsed:   []string{"Go", "PostgreSQL", "Docker", "React"},
				URL:          "",
			},
			{
				Type:         "project",
				Title:        "Open Source CLI Tool",
				Organization: "GitHub",
				StartDate:    "2022-01",
				EndDate:      "2022-06",
				Description:  "Built a CLI tool for database migrations in Go.",
				SkillsUsed:   []string{"Go"},
				URL:          "https://github.com/alice/dbmigrate",
			},
		},
	}
}

func TestBuildPrompt_IncludesAllSkills(t *testing.T) {
	p := testTechProfile()
	prompt := buildTestUserPrompt(p)
	if !strings.Contains(prompt, "Go") {
		t.Fatal("prompt missing Go")
	}
	if !strings.Contains(prompt, "React") {
		t.Fatal("prompt missing React")
	}
	if !strings.Contains(prompt, "Senior Backend Engineer") {
		t.Fatal("prompt missing experience title")
	}
	if !strings.Contains(prompt, "https://github.com") {
		t.Fatal("prompt missing project URL")
	}
}

func testHealthcareProfile() testProfile {
	return testProfile{
		Name:              "Maria Rodriguez",
		Headline:          "Registered Nurse, BSN",
		About:             "Dedicated ER nurse with 12 years of patient care experience.",
		YearsOfExperience: 12,
		Experiences: []testExperience{
			{
				Type:         "employment",
				Title:        "Head Nurse — Emergency Department",
				Organization: "City General Hospital",
				StartDate:    "2019-04",
				EndDate:      "",
				IsCurrent:    true,
				Description:  "Supervise 15 ER nurses, manage patient triage, coordinate with trauma team.",
				Industry:     "Healthcare",
				SkillsUsed:   []string{"Patient Care", "Triage", "Wound Care", "Team Leadership"},
				URL:          "",
			},
			{
				Type:         "employment",
				Title:        "Staff Nurse",
				Organization: "County Medical Center",
				StartDate:    "2014-02",
				EndDate:      "2019-03",
				IsCurrent:    false,
				Description:  "Provided direct patient care in med-surg unit, administered medications, monitored vitals.",
				Industry:     "Healthcare",
				SkillsUsed:   []string{"Patient Care", "Wound Care", "Phlebotomy"},
				URL:          "",
			},
			{
				Type:         "education",
				Title:        "Bachelor of Science in Nursing",
				Organization: "State University",
				StartDate:    "2010-09",
				EndDate:      "2014-05",
				IsCurrent:    false,
				SkillsUsed:   []string{"Patient Care", "Anatomy", "Pharmacology"},
			},
			{
				Type:         "certification",
				Title:        "Registered Nurse License",
				Organization: "State Board of Nursing",
				StartDate:    "2014-07",
				SkillsUsed:   []string{"Patient Care"},
				URL:          "https://example.com/license/RN12345",
			},
			{
				Type:         "certification",
				Title:        "BLS Certification",
				Organization: "American Heart Association",
				StartDate:    "2023-01",
				EndDate:      "2025-01",
				SkillsUsed:   []string{"Patient Care"},
			},
		},
	}
}

func TestHealthcareProfile_PromptContent(t *testing.T) {
	p := testHealthcareProfile()
	prompt := buildTestUserPrompt(p)
	if !strings.Contains(prompt, "Head Nurse") {
		t.Fatal("prompt missing Head Nurse title for role weight classification")
	}
	if !strings.Contains(prompt, "Registered Nurse License") {
		t.Fatal("prompt missing license entry")
	}
	if !strings.Contains(prompt, "Bachelor of Science in Nursing") {
		t.Fatal("prompt missing education entry")
	}
}

func TestTechProfile_FactsComputeCorrectCount(t *testing.T) {
	p := testTechProfile()
	prompt := buildTestUserPrompt(p)

	mock := lib.NewMockLLMClient()
	mock.ResponseFunc = func(system, user string) interface{} {
		return map[string]interface{}{
			"skills": []map[string]interface{}{
				{
					"skill":             "Go",
					"totalYears":        5.5,
					"numRoles":          2,
					"roleWeight":        "senior",
					"educationLevel":    "none",
					"numCertifications": 0,
					"numLicenses":       0,
					"numProjects":       1,
					"numOrganizations":  2,
					"hasUrl":            true,
				},
				{
					"skill":             "React",
					"totalYears":        2.5,
					"numRoles":          1,
					"roleWeight":        "skilled",
					"educationLevel":    "none",
					"numCertifications": 0,
					"numLicenses":       0,
					"numProjects":       0,
					"numOrganizations":  1,
					"hasUrl":            false,
				},
				{
					"skill":             "PostgreSQL",
					"totalYears":        5.5,
					"numRoles":          2,
					"roleWeight":        "senior",
					"educationLevel":    "none",
					"numCertifications": 0,
					"numLicenses":       0,
					"numProjects":       0,
					"numOrganizations":  2,
					"hasUrl":            false,
				},
				{
					"skill":             "Docker",
					"totalYears":        3.5,
					"numRoles":          2,
					"roleWeight":        "skilled",
					"educationLevel":    "none",
					"numCertifications": 0,
					"numLicenses":       0,
					"numProjects":       0,
					"numOrganizations":  2,
					"hasUrl":            false,
				},
			},
			"strengths":  []map[string]interface{}{{"skill": "Go", "score": 90, "note": "Strong backend"}},
			"weaknesses": []map[string]interface{}{},
			"suggestions": []map[string]interface{}{},
		}
	}

	var llmResult struct {
		Skills      []SkillFacts `json:"skills"`
		Strengths   []SkillNote  `json:"strengths"`
		Weaknesses  []SkillNote  `json:"weaknesses"`
		Suggestions []Suggestion `json:"suggestions"`
	}
	if err := mock.Chat(context.Background(), "system", prompt, &llmResult); err != nil {
		t.Fatalf("mock chat: %v", err)
	}

	if len(llmResult.Skills) != 4 {
		t.Fatalf("expected 4 skills, got %d", len(llmResult.Skills))
	}

	counts := make([]SkillCountResult, 0, len(llmResult.Skills))
	for _, f := range llmResult.Skills {
		counts = append(counts, ComputeSkillCount(f))
	}

	totalCount := ComputeTotalCount(counts)
	if totalCount <= 0 {
		t.Fatalf("expected positive totalCount, got %d", totalCount)
	}
	t.Logf("Computed total Count: %d", totalCount)
}

func TestMockLLM_FactExtractionRoundTrip(t *testing.T) {
	mock := lib.NewMockLLMClient()
	var result struct {
		Skills []SkillFacts `json:"skills"`
	}
	err := mock.Chat(context.Background(), "system", "user", &result)
	if err != nil {
		t.Fatalf("mock chat: %v", err)
	}
	if len(result.Skills) == 0 {
		t.Fatal("expected at least one skill")
	}
	if result.Skills[0].TotalYears <= 0 {
		t.Fatal("expected positive TotalYears")
	}
}
