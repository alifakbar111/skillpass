package evaluation

import (
	"testing"
)

func TestComputeSkillCount_ReactSenior(t *testing.T) {
	facts := SkillFacts{
		Skill:             "React",
		TotalYears:        3,
		NumRoles:          2,
		RoleWeight:        RoleSenior,
		EducationLevel:    EduNone,
		NumCertifications: 1,
		NumProjects:       1,
		NumOrganizations:  2,
	}
	result := ComputeSkillCount(facts)
	if result.Count != 150 {
		t.Errorf("expected Count=150, got %d", result.Count)
	}
}

func TestComputeSkillCount_ReactSeniorWithBachelor(t *testing.T) {
	facts := SkillFacts{
		Skill:             "React",
		TotalYears:        3,
		NumRoles:          2,
		RoleWeight:        RoleSenior,
		EducationLevel:    EduBachelor,
		NumCertifications: 1,
		NumProjects:       1,
		NumOrganizations:  2,
	}
	result := ComputeSkillCount(facts)
	if result.Count != 180 {
		t.Errorf("expected Count=180, got %d", result.Count)
	}
}

func TestComputeSkillCount_Nurse(t *testing.T) {
	facts := SkillFacts{
		Skill:            "Patient Care",
		TotalYears:       8,
		NumRoles:         3,
		RoleWeight:       RoleSenior,
		EducationLevel:   EduDiploma,
		NumLicenses:      2,
		NumOrganizations: 3,
	}
	result := ComputeSkillCount(facts)
	if result.Count != 255 {
		t.Errorf("expected Count=255, got %d", result.Count)
	}
}

func TestComputeSkillCount_EntryNoExtras(t *testing.T) {
	facts := SkillFacts{
		Skill:      "Go",
		TotalYears: 0.5,
		NumRoles:   1,
		RoleWeight: RoleEntry,
	}
	result := ComputeSkillCount(facts)
	if result.Count != 30 {
		t.Errorf("expected Count=30, got %d", result.Count)
	}
}

func TestComputeTotalCount_SingleSkill(t *testing.T) {
	skills := []SkillCountResult{
		{Skill: "Go", Count: 150},
	}
	total := ComputeTotalCount(skills)
	if total != 150 {
		t.Errorf("expected total=150, got %d", total)
	}
}

func TestComputeTotalCount_MultiSkill(t *testing.T) {
	skills := []SkillCountResult{
		{Skill: "Go", Count: 150},
		{Skill: "React", Count: 120},
		{Skill: "Docker", Count: 80},
	}
	total := ComputeTotalCount(skills)
	if total != 370 {
		t.Errorf("expected total=370, got %d", total)
	}
}

func TestComputeSkillCount_RoleWeightPoints(t *testing.T) {
	tests := []struct {
		name     string
		weight   RoleWeight
		expected int
	}{
		{"entry", RoleEntry, 10},
		{"skilled", RoleSkilled, 35},
		{"senior", RoleSenior, 65},
		{"expert", RoleExpert, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			facts := SkillFacts{
				Skill:      "test",
				RoleWeight: tt.weight,
			}
			result := ComputeSkillCount(facts)
			if result.Count != tt.expected {
				t.Errorf("RoleWeight=%s: expected Count=%d, got %d", tt.weight, tt.expected, result.Count)
			}
		})
	}
}

func TestComputeSkillCount_EducationPoints(t *testing.T) {
	tests := []struct {
		name     string
		level    EducationLevel
		expected int
	}{
		{"none", EduNone, 0},
		{"hs", EduHS, 5},
		{"diploma", EduDiploma, 15},
		{"bachelor", EduBachelor, 30},
		{"master", EduMaster, 45},
		{"phd", EduPhD, 65},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			facts := SkillFacts{
				Skill:          "test",
				EducationLevel: tt.level,
			}
			result := ComputeSkillCount(facts)
			if result.EducationPoints != tt.expected {
				t.Errorf("EducationLevel=%s: expected EducationPoints=%d, got %d", tt.level, tt.expected, result.EducationPoints)
			}
		})
	}
}
