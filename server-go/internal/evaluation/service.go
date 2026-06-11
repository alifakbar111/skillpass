package evaluation

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"

	"skillpass-server-go/.gen/skillpass/public/model"
	"skillpass-server-go/internal/gen"
	"skillpass-server-go/internal/lib"
)

// Service handles AI evaluation business logic.
type Service struct {
	db  *sql.DB
	llm lib.LLMClient
}

func NewService(db *sql.DB, llm lib.LLMClient) *Service {
	return &Service{db: db, llm: llm}
}

// EvaluationResult is the internal result shape.
type EvaluationResult struct {
	ID           string
	OverallScore int
	Strengths    []SkillNote
	Weaknesses   []SkillNote
	Suggestion   []Suggestion
	SkillScores  []SkillScoreItem
	CreatedAt    string
	RawAnalysis  string
}

type SkillNote struct {
	Skill string `json:"skill"`
	Score int    `json:"score"`
	Note  string `json:"note"`
} //@name SkillNote

type Suggestion struct {
	Area string `json:"area"`
	Tip  string `json:"tip"`
} //@name Suggestion

type SkillScoreItem struct {
	Skill    string `json:"skill"`
	Category string `json:"category"`
	Score    int    `json:"score"`
} //@name SkillScoreItem

func (s *Service) Evaluate(ctx context.Context, profileID string) (*EvaluationResult, error) {
	// 1. Load full profile + experiences
	profileData, err := s.loadFullProfile(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("load profile: %w", err)
	}

	// 2. Build the LLM prompt
	systemPrompt := `You are a career assessment AI. Evaluate the jobseeker's profile and return a JSON object with:
- overallScore (integer, cumulative — every skill, experience, and strength adds points, no upper limit)
- strengths (array of {skill: string, score: int, note: string})
- weaknesses (array of {skill: string, score: int, note: string})
- suggestions (array of {area: string, tip: string})
- skillScores (array of {skill: string, category: string, score: int})

Scoring principles:
- Every skill, year of experience, job entry, and identified strength adds points.
- Weaknesses identify gaps but do NOT subtract from the score.
- No upper limit — encourage honest, complete profiles.
- Categories for skillScores: "backend", "frontend", "devops", "data", "design", "management", "communication", "domain", "tooling"`

	userPrompt := fmt.Sprintf(`Evaluate this jobseeker profile:

Name: %s
Headline: %s
About: %s
Years of Experience: %d

Experience entries:
%s

Skills mentioned across experiences: %s

Return the evaluation as a JSON object following the schema defined in the system prompt.`,
		profileData.Name,
		nullStr(profileData.Headline),
		nullStr(profileData.About),
		nullInt(profileData.YearsOfExperience),
		formatExperiences(profileData.Experiences),
		strings.Join(profileData.AllSkills, ", "))

	// 3. Call LLM
	var llmResult struct {
		OverallScore int             `json:"overallScore"`
		Strengths    []SkillNote     `json:"strengths"`
		Weaknesses   []SkillNote     `json:"weaknesses"`
		Suggestions  []Suggestion    `json:"suggestions"`
		SkillScores  []SkillScoreItem `json:"skillScores"`
	}

	if err := s.llm.Chat(ctx, systemPrompt, userPrompt, &llmResult); err != nil {
		return nil, fmt.Errorf("llm evaluation: %w", err)
	}

	// 4. Marshal the structured parts to JSONB
	strengthsJSON, _ := json.Marshal(llmResult.Strengths)
	weaknessesJSON, _ := json.Marshal(llmResult.Weaknesses)
	suggestionsJSON, _ := json.Marshal(llmResult.Suggestions)
	skillScoresJSON, _ := json.Marshal(llmResult.SkillScores)

	rawAnalysis := fmt.Sprintf("system: %s\n\nuser: %s", systemPrompt, userPrompt)

	// 5. Delete old evaluations for this profile, then insert new one
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Delete previous evaluations
	delStmt := gen.AiEvaluations.DELETE().WHERE(
		gen.AiEvaluations.ProfileID.EQ(UUID(uuid.MustParse(profileID))),
	)
	if _, err := delStmt.ExecContext(ctx, tx); err != nil {
		return nil, fmt.Errorf("delete old evaluations: %w", err)
	}

	// Insert new evaluation
	newID := uuid.New()
	insStmt := gen.AiEvaluations.INSERT(
		gen.AiEvaluations.ID,
		gen.AiEvaluations.ProfileID,
		gen.AiEvaluations.OverallScore,
		gen.AiEvaluations.Strengths,
		gen.AiEvaluations.Weaknesses,
		gen.AiEvaluations.Suggestions,
		gen.AiEvaluations.SkillScores,
		gen.AiEvaluations.RawAnalysis,
	).VALUES(
		newID,
		uuid.MustParse(profileID),
		Int(int64(llmResult.OverallScore)),
		StringExp(CAST(String(string(strengthsJSON))).AS("jsonb")),
		StringExp(CAST(String(string(weaknessesJSON))).AS("jsonb")),
		StringExp(CAST(String(string(suggestionsJSON))).AS("jsonb")),
		StringExp(CAST(String(string(skillScoresJSON))).AS("jsonb")),
		String(rawAnalysis),
	)

	if _, err := insStmt.ExecContext(ctx, tx); err != nil {
		return nil, fmt.Errorf("insert evaluation: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &EvaluationResult{
		ID:           newID.String(),
		OverallScore: llmResult.OverallScore,
		Strengths:    llmResult.Strengths,
		Weaknesses:   llmResult.Weaknesses,
		Suggestion:   llmResult.Suggestions,
		SkillScores:  llmResult.SkillScores,
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
		RawAnalysis:  rawAnalysis,
	}, nil
}

type fullProfile struct {
	Name             string
	Headline         *string
	About            *string
	YearsOfExperience *int32
	Experiences      []model.JobExperiences
	AllSkills        []string
}

func (s *Service) loadFullProfile(ctx context.Context, profileID string) (*fullProfile, error) {
	// Load profile
	stmt := SELECT(
		gen.JobseekerProfiles.ID,
		gen.JobseekerProfiles.Headline,
		gen.JobseekerProfiles.About,
		gen.JobseekerProfiles.YearsOfExperience,
		gen.Users.Name,
	).FROM(
		gen.JobseekerProfiles.INNER_JOIN(gen.Users, gen.Users.ID.EQ(gen.JobseekerProfiles.UserID)),
	).WHERE(
		gen.JobseekerProfiles.ID.EQ(UUID(uuid.MustParse(profileID))),
	)

	var profiles []struct {
		model.JobseekerProfiles
		Name string `alias:"users.name"`
	}
	err := stmt.QueryContext(ctx, s.db, &profiles)
	if err != nil {
		return nil, err
	}
	if len(profiles) == 0 {
		return nil, sql.ErrNoRows
	}
	profile := profiles[0]

	// Load experiences
	expStmt := SELECT(
		gen.JobExperiences.ID,
		gen.JobExperiences.ProfileID,
		gen.JobExperiences.Type,
		gen.JobExperiences.Title,
		gen.JobExperiences.Organization,
		gen.JobExperiences.StartDate,
		gen.JobExperiences.EndDate,
		gen.JobExperiences.IsCurrent,
		gen.JobExperiences.Description,
		gen.JobExperiences.Industry,
		gen.JobExperiences.SkillsUsed,
		gen.JobExperiences.URL,
	).FROM(
		gen.JobExperiences,
	).WHERE(
		gen.JobExperiences.ProfileID.EQ(UUID(uuid.MustParse(profileID))),
	).ORDER_BY(
		gen.JobExperiences.StartDate.ASC(),
	)

	var exps []model.JobExperiences
	if err := expStmt.QueryContext(ctx, s.db, &exps); err != nil {
		return nil, err
	}

	// Collect all unique skills
	skillSet := map[string]struct{}{}
	for _, exp := range exps {
		if exp.SkillsUsed != nil {
			for _, s := range *exp.SkillsUsed {
				skillSet[s] = struct{}{}
			}
		}
	}
	allSkills := make([]string, 0, len(skillSet))
	for s := range skillSet {
		allSkills = append(allSkills, s)
	}

	return &fullProfile{
		Name:              profile.Name,
		Headline:          profile.Headline,
		About:             profile.About,
		YearsOfExperience: profile.YearsOfExperience,
		Experiences:       exps,
		AllSkills:         allSkills,
	}, nil
}

func formatExperiences(exps []model.JobExperiences) string {
	var lines []string
	for _, e := range exps {
		skills := ""
		if e.SkillsUsed != nil {
			skills = strings.Join(*e.SkillsUsed, ", ")
		}
		endDate := e.EndDate
		if endDate == nil || *endDate == "" {
			if e.IsCurrent {
				now := time.Now().Format("2006-01")
				endDate = &now
			}
		}
		line := fmt.Sprintf("- %s at %s (%s to %s) [%s] Skills: %s", e.Title, e.Organization, e.StartDate, nullStr(endDate), string(e.Type), skills)
		if e.Description != nil && *e.Description != "" {
			line += "\n  Description: " + *e.Description
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// GetLatest returns the most recent evaluation for a profile.
func (s *Service) GetLatest(ctx context.Context, profileID string) (*EvaluationResult, error) {
	stmt := SELECT(
		gen.AiEvaluations.AllColumns,
	).FROM(
		gen.AiEvaluations,
	).WHERE(
		gen.AiEvaluations.ProfileID.EQ(UUID(uuid.MustParse(profileID))),
	).ORDER_BY(
		gen.AiEvaluations.CreatedAt.DESC(),
	).LIMIT(1)

	var evals []model.AiEvaluations
	err := stmt.QueryContext(ctx, s.db, &evals)
	if err != nil {
		return nil, err
	}
	if len(evals) == 0 {
		return nil, sql.ErrNoRows
	}
	eval := evals[0]

	var strengths []SkillNote
	var weaknesses []SkillNote
	var suggestions []Suggestion
	var skillScores []SkillScoreItem

	if err := json.Unmarshal([]byte(eval.Strengths), &strengths); err != nil {
		return nil, fmt.Errorf("unmarshal strengths: %w", err)
	}
	if err := json.Unmarshal([]byte(eval.Weaknesses), &weaknesses); err != nil {
		return nil, fmt.Errorf("unmarshal weaknesses: %w", err)
	}
	if err := json.Unmarshal([]byte(eval.Suggestions), &suggestions); err != nil {
		return nil, fmt.Errorf("unmarshal suggestions: %w", err)
	}
	if err := json.Unmarshal([]byte(eval.SkillScores), &skillScores); err != nil {
		return nil, fmt.Errorf("unmarshal skillScores: %w", err)
	}

	return &EvaluationResult{
		ID:           eval.ID.String(),
		OverallScore: int(eval.OverallScore),
		Strengths:    strengths,
		Weaknesses:   weaknesses,
		Suggestion:   suggestions,
		SkillScores:  skillScores,
		CreatedAt:    eval.CreatedAt.Format(time.RFC3339),
		RawAnalysis:  eval.RawAnalysis,
	}, nil
}

// SuggestedRole is one career path recommendation.
type SuggestedRole struct {
	Title     string `json:"title"`
	Reason    string `json:"reason"`
	Readiness string `json:"readiness"` // "ready", "stretch", "long-term"
} //@name SuggestedRole

// DevelopmentStep is a concrete action toward the suggested roles.
type DevelopmentStep struct {
	Area   string `json:"area"`
	Action string `json:"action"`
} //@name DevelopmentStep

// CareerPathResult is the LLM-generated career guidance payload.
type CareerPathResult struct {
	CurrentPosition string            `json:"currentPosition"`
	SuggestedRoles  []SuggestedRole   `json:"suggestedRoles"`
	Steps           []DevelopmentStep `json:"steps"`
} //@name CareerPathResult

// CareerPath asks the LLM for role recommendations based on the profile and its
// latest evaluation. Returns sql.ErrNoRows if no evaluation exists yet.
func (s *Service) CareerPath(ctx context.Context, profileID string) (*CareerPathResult, error) {
	eval, err := s.GetLatest(ctx, profileID)
	if err != nil {
		return nil, err // sql.ErrNoRows passes through for the handler
	}

	profileData, err := s.loadFullProfile(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("load profile: %w", err)
	}

	skillSummary := make([]string, 0, len(eval.SkillScores))
	for _, ss := range eval.SkillScores {
		skillSummary = append(skillSummary, fmt.Sprintf("%s (%s, %d)", ss.Skill, ss.Category, ss.Score))
	}

	systemPrompt := `You are a career advisor AI. Based on the candidate's profile and skill evaluation, return a JSON object with:
- currentPosition (string): a one-line read of where they are in their career
- suggestedRoles (array of exactly 3 items): {title: string, reason: string, readiness: one of "ready", "stretch", "long-term"}
- steps (array of 3-5 items): {area: string, action: string} — concrete, specific development actions

Rules:
- Ground every suggestion in the actual skills and experience provided; do not invent experience.
- "ready" = could apply today; "stretch" = 6-12 months of focused growth; "long-term" = a multi-year trajectory.
- Return ONLY the JSON object.`

	userPrompt := fmt.Sprintf(`Candidate profile:

Headline: %s
Years of experience: %d
Overall evaluation score: %d

Skill scores: %s

Experience entries:
%s

Suggest career paths as JSON per the system prompt schema.`,
		nullStr(profileData.Headline),
		nullInt(profileData.YearsOfExperience),
		eval.OverallScore,
		strings.Join(skillSummary, ", "),
		formatExperiences(profileData.Experiences))

	var result CareerPathResult
	if err := s.llm.Chat(ctx, systemPrompt, userPrompt, &result); err != nil {
		return nil, fmt.Errorf("llm career path: %w", err)
	}

	if result.SuggestedRoles == nil {
		result.SuggestedRoles = []SuggestedRole{}
	}
	if result.Steps == nil {
		result.Steps = []DevelopmentStep{}
	}
	return &result, nil
}

func nullStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func nullInt(v *int32) int {
	if v == nil {
		return 0
	}
	return int(*v)
}
