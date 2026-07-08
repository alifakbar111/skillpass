package matching

import (
	"testing"
)

func TestAssignSkillsToCategory(t *testing.T) {
	mapping := AssignSkillsToCategory([]string{"Go", "React", "SQL", "AWS", "Patient Care"})
	if mapping["Go"] != "Software Engineering" {
		t.Fatalf("Go: expected Software Engineering, got %s", mapping["Go"])
	}
	if mapping["SQL"] != "Data & Analytics" {
		t.Fatalf("SQL: expected Data & Analytics, got %s", mapping["SQL"])
	}
	if mapping["Patient Care"] != "Clinical & Medical" {
		t.Fatalf("Patient Care: expected Clinical & Medical, got %s", mapping["Patient Care"])
	}
}

func TestAssignSkillsToCategory_Default(t *testing.T) {
	mapping := AssignSkillsToCategory([]string{"UnknownSkill"})
	if mapping["UnknownSkill"] != "Software Engineering" {
		t.Fatalf("expected default Software Engineering, got %s", mapping["UnknownSkill"])
	}
}
