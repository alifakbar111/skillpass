package career

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"skillpass-server-go/internal/lib"
)

// sanitizeLLMInput strips newlines and limits length to prevent prompt injection.
func sanitizeLLMInput(input string, maxLen int) string {
	input = strings.ReplaceAll(input, "\n", " ")
	input = strings.ReplaceAll(input, "\r", " ")
	input = strings.TrimSpace(input)
	if len(input) > maxLen {
		input = input[:maxLen]
	}
	return input
}

type Service struct {
	db  *sql.DB
	llm lib.LLMClient
}

func NewService(db *sql.DB, llm lib.LLMClient) *Service {
	return &Service{db: db, llm: llm}
}

type SkillRequirement struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Level    string `json:"level,omitempty"`
}

type ProgressionStep struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Duration    string `json:"duration,omitempty"`
}

type CareerPath struct {
	ID                  string               `json:"id"`
	Title               string               `json:"title"`
	Description         string               `json:"description"`
	SkillRequirements   []SkillRequirement   `json:"skillRequirements"`
	TypicalProgression  []ProgressionStep    `json:"typicalProgression"`
	Industry            string               `json:"industry"`
	CreatedAt           string               `json:"createdAt"`
}

type SkillGapItem struct {
	Skill          string `json:"skill"`
	Required       bool   `json:"required"`
	RequiredLevel  string `json:"requiredLevel,omitempty"`
	UserLevel      int    `json:"userLevel"`
	Gap            int    `json:"gap"`
}

type SkillGapResult struct {
	ProfileID       string          `json:"profileId"`
	MatchingSkills  []SkillGapItem  `json:"matchingSkills"`
	MissingSkills   []SkillGapItem  `json:"missingSkills"`
	OverallMatch    float64         `json:"overallMatch"`
}

type CareerPrediction struct {
	CurrentPosition    string              `json:"currentPosition"`
	PredictedPaths     []PredictedPath     `json:"predictedPaths"`
	SkillDevelopment   []SkillDevelopment  `json:"skillDevelopment"`
	EstimatedTimeline  string              `json:"estimatedTimeline"`
}

type PredictedPath struct {
	Title       string  `json:"title"`
	Confidence  float64 `json:"confidence"`
	Reasoning   string  `json:"reasoning"`
}

type SkillDevelopment struct {
	Skill       string `json:"skill"`
	Current     int    `json:"current"`
	Target      int    `json:"target"`
	Actions     []string `json:"actions"`
}

func (s *Service) ListPaths(ctx context.Context, industry string) ([]CareerPath, error) {
	query := `
		SELECT id, title, description, skill_requirements, typical_progression, industry, created_at
		FROM career_paths
		WHERE ($1 = '' OR industry = $1)
		ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, industry)
	if err != nil {
		return nil, fmt.Errorf("query career paths: %w", err)
	}
	defer rows.Close()

	var paths []CareerPath
	for rows.Next() {
		var p CareerPath
		var id uuid.UUID
		var skillReqJSON, progressionJSON string
		var createdAt interface{}

		if err := rows.Scan(&id, &p.Title, &p.Description, &skillReqJSON, &progressionJSON, &p.Industry, &createdAt); err != nil {
			return nil, fmt.Errorf("scan career path: %w", err)
		}

		p.ID = id.String()

		if err := json.Unmarshal([]byte(skillReqJSON), &p.SkillRequirements); err != nil {
			p.SkillRequirements = []SkillRequirement{}
		}
		if err := json.Unmarshal([]byte(progressionJSON), &p.TypicalProgression); err != nil {
			p.TypicalProgression = []ProgressionStep{}
		}

		if t, ok := createdAt.(interface{ Format(string) string }); ok {
			p.CreatedAt = t.Format("2006-01-02T15:04:05Z07:00")
		}

		paths = append(paths, p)
	}

	if paths == nil {
		paths = []CareerPath{}
	}

	return paths, nil
}

func (s *Service) GetSkillGap(ctx context.Context, profileID string, industry string) (*SkillGapResult, error) {
	// Get career paths for the industry
	pathQuery := `
		SELECT skill_requirements
		FROM career_paths
		WHERE industry = $1
	`

	rows, err := s.db.QueryContext(ctx, pathQuery, industry)
	if err != nil {
		return nil, fmt.Errorf("query career paths: %w", err)
	}
	defer rows.Close()

	requiredSkills := map[string]SkillRequirement{}
	for rows.Next() {
		var skillReqJSON string
		if err := rows.Scan(&skillReqJSON); err != nil {
			continue
		}

		var reqs []SkillRequirement
		if err := json.Unmarshal([]byte(skillReqJSON), &reqs); err != nil {
			continue
		}

		for _, req := range reqs {
			if _, ok := requiredSkills[req.Name]; !ok || req.Required {
				requiredSkills[req.Name] = req
			}
		}
	}

	// Get user's skill scores from latest evaluation
	evalQuery := `
		SELECT skill_scores
		FROM ai_evaluations
		WHERE profile_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var skillScoresJSON string
	err = s.db.QueryRowContext(ctx, evalQuery, profileID).Scan(&skillScoresJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return &SkillGapResult{
				ProfileID:      profileID,
				MatchingSkills: []SkillGapItem{},
				MissingSkills:  []SkillGapItem{},
				OverallMatch:   0,
			}, nil
		}
		return nil, fmt.Errorf("query evaluation: %w", err)
	}

	type SkillScoreItem struct {
		Skill string `json:"skill"`
		Score int    `json:"score"`
	}

	var userScores []SkillScoreItem
	if err := json.Unmarshal([]byte(skillScoresJSON), &userScores); err != nil {
		return nil, fmt.Errorf("unmarshal skill scores: %w", err)
	}

	userSkillMap := map[string]int{}
	for _, s := range userScores {
		userSkillMap[strings.ToLower(s.Skill)] = s.Score
	}

	var matching []SkillGapItem
	var missing []SkillGapItem

	for skillName, req := range requiredSkills {
		userScore, hasSkill := userSkillMap[strings.ToLower(skillName)]

		if hasSkill {
			matching = append(matching, SkillGapItem{
				Skill:         skillName,
				Required:      req.Required,
				RequiredLevel: req.Level,
				UserLevel:     userScore,
				Gap:           0,
			})
		} else {
			missing = append(missing, SkillGapItem{
				Skill:         skillName,
				Required:      req.Required,
				RequiredLevel: req.Level,
				UserLevel:     0,
				Gap:           100,
			})
		}
	}

	totalSkills := len(requiredSkills)
	matchedCount := len(matching)
	overallMatch := 0.0
	if totalSkills > 0 {
		overallMatch = float64(matchedCount) / float64(totalSkills) * 100
	}

	if matching == nil {
		matching = []SkillGapItem{}
	}
	if missing == nil {
		missing = []SkillGapItem{}
	}

	return &SkillGapResult{
		ProfileID:      profileID,
		MatchingSkills: matching,
		MissingSkills:  missing,
		OverallMatch:   overallMatch,
	}, nil
}

func (s *Service) PredictPath(ctx context.Context, profileID string, industry string) (*CareerPrediction, error) {
	// Get user's skill scores
	evalQuery := `
		SELECT skill_scores
		FROM ai_evaluations
		WHERE profile_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	type SkillScoreItem struct {
		Skill    string `json:"skill"`
		Category string `json:"category"`
		Score    int    `json:"score"`
	}

	var skillScoresJSON string
	err := s.db.QueryRowContext(ctx, evalQuery, profileID).Scan(&skillScoresJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return &CareerPrediction{
				CurrentPosition:   "Unknown",
				PredictedPaths:    []PredictedPath{},
				SkillDevelopment:  []SkillDevelopment{},
				EstimatedTimeline: "Unable to determine",
			}, nil
		}
		return nil, fmt.Errorf("query evaluation: %w", err)
	}

	var userScores []SkillScoreItem
	if err := json.Unmarshal([]byte(skillScoresJSON), &userScores); err != nil {
		return nil, fmt.Errorf("unmarshal skill scores: %w", err)
	}

	// Get profile info
	profileQuery := `
		SELECT jsp.headline, jsp.years_of_experience, u.name
		FROM jobseeker_profiles jsp
		JOIN users u ON u.id = jsp.user_id
		WHERE jsp.id = $1
	`

	var headline sql.NullString
	var yearsExp sql.NullInt32
	var name string
	if err := s.db.QueryRowContext(ctx, profileQuery, profileID).Scan(&headline, &yearsExp, &name); err != nil {
		return nil, fmt.Errorf("query profile: %w", err)
	}

	// Build skill summary
	skillSummary := make([]string, 0, len(userScores))
	for _, ss := range userScores {
		skillSummary = append(skillSummary, fmt.Sprintf("%s (%s, %d)", ss.Skill, ss.Category, ss.Score))
	}

	headlineStr := ""
	if headline.Valid {
		headlineStr = headline.String
	}

	yearsExpInt := 0
	if yearsExp.Valid {
		yearsExpInt = int(yearsExp.Int32)
	}

	systemPrompt := `You are a career prediction AI. Based on the candidate's profile and skill evaluation, predict their career trajectory.
Return a JSON object with:
- currentPosition (string): where they are now
- predictedPaths (array of 2-3): {title: string, confidence: 0.0-1.0, reasoning: string}
- skillDevelopment (array of 2-4): {skill: string, current: int, target: int, actions: array of strings}
- estimatedTimeline (string): how long to reach next level

Rules:
- Ground predictions in actual skills and experience
- Confidence should reflect how well their skills match the predicted path
- Provide specific, actionable development steps`

	userPrompt := fmt.Sprintf(`Candidate: %s
Headline: %s
Years of experience: %d
Industry: %s
Skill scores: %s

Predict career paths as JSON per the system prompt schema.`, sanitizeLLMInput(name, 200), sanitizeLLMInput(headlineStr, 200), yearsExpInt, sanitizeLLMInput(industry, 100), strings.Join(skillSummary, ", "))

	var prediction CareerPrediction
	if err := s.llm.Chat(ctx, systemPrompt, userPrompt, &prediction); err != nil {
		return nil, fmt.Errorf("llm prediction: %w", err)
	}

	if prediction.PredictedPaths == nil {
		prediction.PredictedPaths = []PredictedPath{}
	}
	if prediction.SkillDevelopment == nil {
		prediction.SkillDevelopment = []SkillDevelopment{}
	}

	return &prediction, nil
}
