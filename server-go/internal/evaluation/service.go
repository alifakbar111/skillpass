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
	"github.com/uptrace/bun"

	"skillpass-server-go/.gen/skillpass/public/model"
	"skillpass-server-go/internal/gen"
	"skillpass-server-go/internal/lib"
)

const (
	EvaluationValidDuration = 3 * 30 * 24 * time.Hour // ~3 months
)

// IsExpired returns true if the evaluation was created more than 3 months ago.
func IsExpired(createdAt time.Time) bool {
	return time.Since(createdAt) > EvaluationValidDuration
}

// Service handles AI evaluation business logic.
type Service struct {
	db  *sql.DB
	bun bun.IDB
	llm lib.LLMClient
}

func NewService(db *sql.DB, llm lib.LLMClient, bun bun.IDB) *Service {
	return &Service{db: db, bun: bun, llm: llm}
}

// EvaluationResult is the internal result shape.
type EvaluationResult struct {
	ID           string
	OverallScore int
	Strengths    []SkillNote
	Weaknesses   []SkillNote
	Suggestion   []Suggestion
	SkillScores  []SkillScoreItem
	SkillCounts  []SkillCountResult
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
	profileUUID, err := lib.ParseUUID(profileID)
	if err != nil {
		return nil, fmt.Errorf("invalid profile ID: %w", err)
	}

	// 1. Load full profile + experiences
	profileData, err := s.loadFullProfile(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("load profile: %w", err)
	}

	// 2. Build the LLM prompt for fact extraction
	systemPrompt := `You are a career assessment data extractor. For each skill found in the profile, extract the following structured facts. Do NOT calculate scores — just extract facts.

For EACH skill, extract:
1. totalYears — Total calendar years the skill was actively used. If multiple roles overlapped in time, do NOT double-count — use the actual calendar duration.
2. numRoles — Number of distinct roles/experiences that mention this skill
3. roleWeight — Highest role weight bucket across all roles:
   - "entry": Junior/Intern/Associate titles, basic or routine tasks, IC scope
   - "skilled": Mid-level, independent work, moderate complexity
   - "senior": Senior/Lead/Supervisor/Head, complex non-routine work, leads others
   - "expert": Manager/Director/VP/C-level, org-wide/strategic impact
4. educationLevel — Highest education level studied for this skill:
   - "none", "hs" (high school), "diploma", "bachelor", "master", "phd"
5. numCertifications — Count of standard certification entries matching this skill (third-party, company, org)
6. numLicenses — Count of professional license entries matching this skill (RN, CPA, PE, etc.)
7. numProjects — Number of project/portfolio entries using this skill
8. numOrganizations — Number of distinct organizations where this skill was used
9. hasUrl — Does any experience entry have a URL showing work with this skill?

Also generate:
- strengths (array of {skill: string, score: int, note: string})
- weaknesses (array of {skill: string, score: int, note: string})
- suggestions (array of {area: string, tip: string})

Do NOT return an overallScore or skillScores array. This is fact extraction only.
Return ONLY valid JSON.`

	userPrompt := fmt.Sprintf(`Extract skill facts from this jobseeker profile:

Name: %s
Headline: %s
About: %s
Years of Experience: %d

Experience entries:
%s

Skills mentioned across experiences: %s

Return the extracted facts as JSON per the system prompt schema.`,
		profileData.Name,
		nullStr(profileData.Headline),
		nullStr(profileData.About),
		nullInt(profileData.YearsOfExperience),
		formatExperiences(profileData.Experiences),
		strings.Join(profileData.AllSkills, ", "))

	// 3. Call LLM — expect facts, not scores
	var llmResult struct {
		Skills      []SkillFacts `json:"skills"`
		Strengths   []SkillNote  `json:"strengths"`
		Weaknesses  []SkillNote  `json:"weaknesses"`
		Suggestions []Suggestion `json:"suggestions"`
	}

	if err := s.llm.Chat(ctx, systemPrompt, userPrompt, &llmResult); err != nil {
		return nil, fmt.Errorf("llm evaluation: %w", err)
	}

	// 4. Compute Counts server-side from extracted facts
	skillCounts := make([]SkillCountResult, 0, len(llmResult.Skills))
	skillScores := make([]SkillScoreItem, 0, len(llmResult.Skills))
	for _, facts := range llmResult.Skills {
		result := ComputeSkillCount(facts)
		skillCounts = append(skillCounts, result)
		skillScores = append(skillScores, SkillScoreItem{
			Skill: result.Skill,
			Score: result.Count,
		})
	}

	totalCount := ComputeTotalCount(skillCounts)

	// 5. Marshal to JSONB
	strengthsJSON, err := json.Marshal(llmResult.Strengths)
	if err != nil {
		return nil, fmt.Errorf("marshal strengths: %w", err)
	}
	weaknessesJSON, err := json.Marshal(llmResult.Weaknesses)
	if err != nil {
		return nil, fmt.Errorf("marshal weaknesses: %w", err)
	}
	suggestionsJSON, err := json.Marshal(llmResult.Suggestions)
	if err != nil {
		return nil, fmt.Errorf("marshal suggestions: %w", err)
	}
	skillScoresJSON, err := json.Marshal(skillScores)
	if err != nil {
		return nil, fmt.Errorf("marshal skillScores: %w", err)
	}
	rawAnalysis := fmt.Sprintf("system: %s\n\nuser: %s\n\nllm_facts: %s",
		systemPrompt, userPrompt, mustMarshal(llmResult.Skills))

	// 6. Insert evaluation with is_current lifecycle management
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Flag all existing current evaluations as not current
	updateStmt := gen.AiEvaluations.
		UPDATE(gen.AiEvaluations.IsCurrent).
		SET(Bool(false)).
		WHERE(
			gen.AiEvaluations.ProfileID.EQ(UUID(profileUUID)).
				AND(gen.AiEvaluations.IsCurrent.EQ(Bool(true))),
		)
	if _, err := updateStmt.ExecContext(ctx, tx); err != nil {
		return nil, fmt.Errorf("flag old evaluations: %w", err)
	}

	// Insert new evaluation with is_current = true
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
		gen.AiEvaluations.IsCurrent,
	).VALUES(
		newID,
		profileUUID,
		Int(int64(totalCount)),
		StringExp(CAST(String(string(strengthsJSON))).AS("jsonb")),
		StringExp(CAST(String(string(weaknessesJSON))).AS("jsonb")),
		StringExp(CAST(String(string(suggestionsJSON))).AS("jsonb")),
		StringExp(CAST(String(string(skillScoresJSON))).AS("jsonb")),
		String(rawAnalysis),
		Bool(true),
	)

	if _, err := insStmt.ExecContext(ctx, tx); err != nil {
		return nil, fmt.Errorf("insert evaluation: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &EvaluationResult{
		ID:           newID.String(),
		OverallScore: totalCount,
		Strengths:    llmResult.Strengths,
		Weaknesses:   llmResult.Weaknesses,
		Suggestion:   llmResult.Suggestions,
		SkillScores:  skillScores,
		SkillCounts:  skillCounts,
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
		RawAnalysis:  rawAnalysis,
	}, nil
}

// mustMarshal is a helper that returns JSON string or "{}" on error.
func mustMarshal(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
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
	profileUUID, err := lib.ParseUUID(profileID)
	if err != nil {
		return nil, fmt.Errorf("invalid profile ID: %w", err)
	}

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
		gen.JobseekerProfiles.ID.EQ(UUID(profileUUID)),
	)

	var profiles []struct {
		model.JobseekerProfiles
		Name string `alias:"users.name"`
	}
	err = stmt.QueryContext(ctx, s.db, &profiles)
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
		gen.JobExperiences.ProfileID.EQ(UUID(profileUUID)),
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

// GetLatest returns the most recent current evaluation for a profile.
func (s *Service) GetLatest(ctx context.Context, profileID string) (*EvaluationResult, error) {
	profileUUID, err := lib.ParseUUID(profileID)
	if err != nil {
		return nil, fmt.Errorf("invalid profile ID: %w", err)
	}

	stmt := SELECT(
		gen.AiEvaluations.AllColumns,
	).FROM(
		gen.AiEvaluations,
	).WHERE(
		gen.AiEvaluations.ProfileID.EQ(UUID(profileUUID)).
			AND(gen.AiEvaluations.IsCurrent.EQ(Bool(true))),
	).LIMIT(1)

	var evals []model.AiEvaluations
	if err := stmt.QueryContext(ctx, s.db, &evals); err != nil {
		return nil, err
	}
	if len(evals) == 0 {
		return nil, sql.ErrNoRows
	}
	return evalToResult(evals[0]), nil
}

// GetHistory returns all evaluations for a profile, ordered by newest first.
func (s *Service) GetHistory(ctx context.Context, profileID string) ([]*EvaluationResult, error) {
	profileUUID, err := lib.ParseUUID(profileID)
	if err != nil {
		return nil, fmt.Errorf("invalid profile ID: %w", err)
	}

	stmt := SELECT(
		gen.AiEvaluations.AllColumns,
	).FROM(
		gen.AiEvaluations,
	).WHERE(
		gen.AiEvaluations.ProfileID.EQ(UUID(profileUUID)),
	).ORDER_BY(
		gen.AiEvaluations.CreatedAt.DESC(),
	)

	var evals []model.AiEvaluations
	if err := stmt.QueryContext(ctx, s.db, &evals); err != nil {
		return nil, err
	}

	results := make([]*EvaluationResult, 0, len(evals))
	for _, eval := range evals {
		result := evalToResult(eval)
		if result != nil {
			results = append(results, result)
		}
	}
	return results, nil
}

// evalToResult converts a model.AiEvaluations row to EvaluationResult.
func evalToResult(eval model.AiEvaluations) *EvaluationResult {
	var strengths []SkillNote
	var weaknesses []SkillNote
	var suggestions []Suggestion
	var skillScores []SkillScoreItem

	if err := json.Unmarshal([]byte(eval.Strengths), &strengths); err != nil {
		return nil
	}
	if err := json.Unmarshal([]byte(eval.Weaknesses), &weaknesses); err != nil {
		return nil
	}
	if err := json.Unmarshal([]byte(eval.Suggestions), &suggestions); err != nil {
		return nil
	}
	if err := json.Unmarshal([]byte(eval.SkillScores), &skillScores); err != nil {
		return nil
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
	}
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
	SuggestedRoles  []SuggestedRole   `json:"suggestedRoles,omitempty"`
	Steps           []DevelopmentStep `json:"steps,omitempty"`
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
		if ss.Category != "" {
			skillSummary = append(skillSummary, fmt.Sprintf("%s (%s, %d)", ss.Skill, ss.Category, ss.Score))
		} else {
			skillSummary = append(skillSummary, fmt.Sprintf("%s (%d)", ss.Skill, ss.Score))
		}
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
