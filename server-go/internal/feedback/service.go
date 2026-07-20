package feedback

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"skillpass-server-go/internal/lib"
)

type Service struct {
	db      *sql.DB
	llm     lib.LLMClient
	aiSlots chan struct{} // bounded semaphore for in-flight AI generations
}

const defaultMaxInflightAI = 8

func NewService(db *sql.DB, llm lib.LLMClient) *Service {
	return &Service{
		db:      db,
		llm:     llm,
		aiSlots: make(chan struct{}, defaultMaxInflightAI),
	}
}

type RatingArea struct {
	Area   string `json:"area"`
	Score  int    `json:"score"`
	Notes  string `json:"notes"`
} //@name RatingArea

type AISuggestion struct {
	Area   string `json:"area"`
	Tip    string `json:"tip"`
} //@name AISuggestion

type Feedback struct {
	ID             string          `json:"id"`
	ProfileID      string          `json:"profileId"`
	CompanyID      string          `json:"companyId"`
	Content        string          `json:"content"`
	RatingAreas    []RatingArea    `json:"ratingAreas,omitempty"`
	AISuggestions  []AISuggestion  `json:"aiSuggestions,omitempty"`
	CreatedAt      string          `json:"createdAt"`
} //@name Feedback

type CreateFeedbackRequest struct {
	Content     string       `json:"content"`
	RatingAreas []RatingArea `json:"ratingAreas,omitempty"`
} //@name CreateFeedbackRequest

func (s *Service) Create(ctx context.Context, profileID, companyID string, req *CreateFeedbackRequest) (*Feedback, error) {
	id := uuid.New()
	now := time.Now().UTC()

	ratingJSON, err := json.Marshal(req.RatingAreas)
	if err != nil {
		return nil, fmt.Errorf("marshal rating areas: %w", err)
	}

	var feedbackID uuid.UUID
	err = s.db.QueryRowContext(ctx,
		`INSERT INTO feedback (id, profile_id, company_id, content, rating_areas, ai_suggestions, created_at)
		 VALUES ($1, $2, $3, $4, $5, '[]', $6)
		 RETURNING id`,
		id, profileID, companyID, req.Content, string(ratingJSON), now,
	).Scan(&feedbackID)
	if err != nil {
		return nil, fmt.Errorf("insert feedback: %w", err)
	}

	fb := &Feedback{
		ID:            feedbackID.String(),
		ProfileID:     profileID,
		CompanyID:     companyID,
		Content:       req.Content,
		RatingAreas:   req.RatingAreas,
		AISuggestions: []AISuggestion{},
		CreatedAt:     now.Format(time.RFC3339),
	}

	select {
	case s.aiSlots <- struct{}{}:
	default:
		slog.Warn("AI suggestion slot full; skipping background generation", "feedbackID", feedbackID)
		return fb, nil
	}
	go func(id, profile string) {
		defer func() { <-s.aiSlots }()
		s.generateAISuggestions(id, profile)
	}(feedbackID.String(), profileID)

	return fb, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*Feedback, error) {
	var fb Feedback
	var ratingJSON, suggestionJSON string

	err := s.db.QueryRowContext(ctx,
		`SELECT id, profile_id, company_id, content, rating_areas, ai_suggestions, created_at
		 FROM feedback WHERE id = $1`, id,
	).Scan(&fb.ID, &fb.ProfileID, &fb.CompanyID, &fb.Content, &ratingJSON, &suggestionJSON, &fb.CreatedAt)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(ratingJSON), &fb.RatingAreas); err != nil {
		return nil, fmt.Errorf("unmarshal rating areas: %w", err)
	}
	if err := json.Unmarshal([]byte(suggestionJSON), &fb.AISuggestions); err != nil {
		return nil, fmt.Errorf("unmarshal ai suggestions: %w", err)
	}

	return &fb, nil
}

func (s *Service) GetByProfileID(ctx context.Context, profileID string) ([]Feedback, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, profile_id, company_id, content, rating_areas, ai_suggestions, created_at
		 FROM feedback WHERE profile_id = $1 ORDER BY created_at DESC`, profileID,
	)
	if err != nil {
		return nil, fmt.Errorf("query feedback: %w", err)
	}
	defer rows.Close()

	feedbacks := []Feedback{}
	for rows.Next() {
		var fb Feedback
		var ratingJSON, suggestionJSON string

		if err := rows.Scan(&fb.ID, &fb.ProfileID, &fb.CompanyID, &fb.Content, &ratingJSON, &suggestionJSON, &fb.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan feedback: %w", err)
		}
		if err := json.Unmarshal([]byte(ratingJSON), &fb.RatingAreas); err != nil {
			return nil, fmt.Errorf("unmarshal rating areas: %w", err)
		}
		if err := json.Unmarshal([]byte(suggestionJSON), &fb.AISuggestions); err != nil {
			return nil, fmt.Errorf("unmarshal ai suggestions: %w", err)
		}
		feedbacks = append(feedbacks, fb)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate feedback rows: %w", err)
	}
	return feedbacks, nil
}

func (s *Service) GetByCompanyID(ctx context.Context, companyID string) ([]Feedback, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, profile_id, company_id, content, rating_areas, ai_suggestions, created_at
		 FROM feedback WHERE company_id = $1 ORDER BY created_at DESC`, companyID,
	)
	if err != nil {
		return nil, fmt.Errorf("query feedback: %w", err)
	}
	defer rows.Close()

	feedbacks := []Feedback{}
	for rows.Next() {
		var fb Feedback
		var ratingJSON, suggestionJSON string

		if err := rows.Scan(&fb.ID, &fb.ProfileID, &fb.CompanyID, &fb.Content, &ratingJSON, &suggestionJSON, &fb.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan feedback: %w", err)
		}
		if err := json.Unmarshal([]byte(ratingJSON), &fb.RatingAreas); err != nil {
			return nil, fmt.Errorf("unmarshal rating areas: %w", err)
		}
		if err := json.Unmarshal([]byte(suggestionJSON), &fb.AISuggestions); err != nil {
			return nil, fmt.Errorf("unmarshal ai suggestions: %w", err)
		}
		feedbacks = append(feedbacks, fb)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate feedback rows: %w", err)
	}
	return feedbacks, nil
}

func (s *Service) GetAISuggestionsByProfileID(ctx context.Context, profileID string) ([]AISuggestion, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT ai_suggestions FROM feedback
		 WHERE profile_id = $1 AND ai_suggestions != '[]'
		 ORDER BY created_at DESC`, profileID,
	)
	if err != nil {
		return nil, fmt.Errorf("query ai suggestions: %w", err)
	}
	defer rows.Close()

	var all []AISuggestion
	for rows.Next() {
		var suggestionJSON string
		if err := rows.Scan(&suggestionJSON); err != nil {
			return nil, fmt.Errorf("scan suggestion: %w", err)
		}
		var suggestions []AISuggestion
		if err := json.Unmarshal([]byte(suggestionJSON), &suggestions); err != nil {
			return nil, fmt.Errorf("unmarshal suggestions: %w", err)
		}
		all = append(all, suggestions...)
	}
	return all, nil
}

func (s *Service) generateAISuggestions(feedbackID, profileID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var profileName string
	err := s.db.QueryRowContext(ctx,
		`SELECT u.name FROM jobseeker_profiles jp
		 JOIN users u ON u.id = jp.user_id
		 WHERE jp.id = $1`, profileID,
	).Scan(&profileName)
	if err != nil {
		slog.Error("failed to load profile for AI suggestions", "feedbackID", feedbackID, "error", err)
		return
	}

	var feedbackContent string
	var ratingJSON string
	err = s.db.QueryRowContext(ctx,
		`SELECT content, rating_areas FROM feedback WHERE id = $1`, feedbackID,
	).Scan(&feedbackContent, &ratingJSON)
	if err != nil {
		slog.Error("failed to load feedback for AI suggestions", "feedbackID", feedbackID, "error", err)
		return
	}

	var ratings []RatingArea
	if err := json.Unmarshal([]byte(ratingJSON), &ratings); err != nil {
		slog.Error("failed to unmarshal ratings for AI suggestions", "feedbackID", feedbackID, "error", err)
		return
	}

	ratingSummary := ""
	for _, r := range ratings {
		ratingSummary += fmt.Sprintf("- %s: %d/5 — %s\n", r.Area, r.Score, r.Notes)
	}

	systemPrompt := `You are a career development AI. Based on the company's feedback and ratings for a jobseeker, generate actionable improvement suggestions.

Return a JSON array of objects with:
- area (string): the skill or competency area
- tip (string): a specific, actionable improvement suggestion

Generate 3-5 suggestions based on the feedback and ratings. Focus on areas with lower scores.`

	userPrompt := fmt.Sprintf(`Jobseeker: %s

Company feedback:
%s

Rating areas:
%s

Return improvement suggestions as a JSON array.`, profileName, feedbackContent, ratingSummary)

	var result []AISuggestion
	if err := s.llm.Chat(ctx, systemPrompt, userPrompt, &result); err != nil {
		slog.Error("LLM AI suggestions failed", "feedbackID", feedbackID, "error", err)
		return
	}

	suggestionJSON, err := json.Marshal(result)
	if err != nil {
		slog.Error("failed to marshal AI suggestions", "feedbackID", feedbackID, "error", err)
		return
	}

	_, err = s.db.ExecContext(ctx,
		`UPDATE feedback SET ai_suggestions = $1 WHERE id = $2`,
		string(suggestionJSON), feedbackID,
	)
	if err != nil {
		slog.Error("failed to save AI suggestions", "feedbackID", feedbackID, "error", err)
		return
	}

	slog.Info("AI suggestions generated", "feedbackID", feedbackID, "count", len(result))
}
