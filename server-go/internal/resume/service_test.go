package resume

import (
	"context"
	"testing"

	"skillpass-server-go/internal/lib"
)

func TestParseResume(t *testing.T) {
	mock := lib.NewMockLLMClient()
	mock.ResponseFunc = func(_, _ string) interface{} {
		return map[string]interface{}{
			"headline":          "Senior Backend Engineer",
			"about":             "Experienced Go developer.",
			"yearsOfExperience": 6,
			"experiences": []map[string]interface{}{
				{
					"type":         "employment",
					"title":        "Backend Engineer",
					"organization": "Acme Corp",
					"startDate":    "2019-01",
					"endDate":      "2023-06",
					"isCurrent":    false,
					"description":  "Built APIs",
					"skillsUsed":   []string{"Go", "PostgreSQL"},
				},
			},
		}
	}

	svc := NewService(mock)
	result, err := svc.Parse(context.Background(), "long enough resume text for parsing here ......")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if result.Headline != "Senior Backend Engineer" {
		t.Fatalf("unexpected headline %q", result.Headline)
	}
	if result.YearsOfExperience != 6 {
		t.Fatalf("unexpected years %d", result.YearsOfExperience)
	}
	if len(result.Experiences) != 1 {
		t.Fatalf("expected 1 experience, got %d", len(result.Experiences))
	}
	if result.Experiences[0].Title != "Backend Engineer" {
		t.Fatalf("unexpected title %q", result.Experiences[0].Title)
	}
	if len(result.Experiences[0].SkillsUsed) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(result.Experiences[0].SkillsUsed))
	}
}

func TestParseResumeEmptyExperiences(t *testing.T) {
	mock := lib.NewMockLLMClient()
	mock.ResponseFunc = func(_, _ string) interface{} {
		return map[string]interface{}{
			"headline":          "Junior Dev",
			"about":             "",
			"yearsOfExperience": 0,
		}
	}

	svc := NewService(mock)
	result, err := svc.Parse(context.Background(), "some resume text that is long enough to be parsed ok")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if result.Experiences == nil {
		t.Fatal("expected non-nil experiences slice")
	}
	if len(result.Experiences) != 0 {
		t.Fatalf("expected 0 experiences, got %d", len(result.Experiences))
	}
}
