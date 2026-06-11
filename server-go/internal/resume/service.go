package resume

import (
	"context"
	"fmt"

	"skillpass-server-go/internal/lib"
)

// Service parses raw resume text into structured profile data using an LLM.
type Service struct {
	llm lib.LLMClient
}

func NewService(llm lib.LLMClient) *Service {
	return &Service{llm: llm}
}

// ParsedExperience mirrors the experience entry shape the frontend can save
// via POST /profiles/me/experience.
type ParsedExperience struct {
	Type         string   `json:"type"`
	Title        string   `json:"title"`
	Organization string   `json:"organization"`
	StartDate    string   `json:"startDate"`
	EndDate      string   `json:"endDate"`
	IsCurrent    bool     `json:"isCurrent"`
	Description  string   `json:"description"`
	SkillsUsed   []string `json:"skillsUsed"`
}

// ParsedResume is the structured extraction returned to the client for review.
type ParsedResume struct {
	Headline          string             `json:"headline"`
	About             string             `json:"about"`
	YearsOfExperience int                `json:"yearsOfExperience"`
	Experiences       []ParsedExperience `json:"experiences"`
}

func (s *Service) Parse(ctx context.Context, resumeText string) (*ParsedResume, error) {
	systemPrompt := `You are a resume parser. Extract structured career data from the resume text and return a JSON object with:
- headline (string): a concise professional headline (e.g. "Senior Backend Engineer")
- about (string): a 1-2 sentence professional summary
- yearsOfExperience (integer): total years of professional experience, estimated from the entries
- experiences (array): each item has:
  - type: one of "employment", "gig", "education", "certification", "project", "volunteering"
  - title (string)
  - organization (string)
  - startDate (string, format "YYYY-MM")
  - endDate (string, format "YYYY-MM", empty string if current/ongoing)
  - isCurrent (boolean)
  - description (string)
  - skillsUsed (array of strings)

Rules:
- Map jobs to "employment", schools/degrees to "education", certs to "certification", side projects to "project".
- If a date is unknown, use a reasonable estimate in YYYY-MM form; never invent organizations.
- Return ONLY the JSON object.`

	userPrompt := fmt.Sprintf("Parse this resume:\n\n%s", resumeText)

	var result ParsedResume
	if err := s.llm.Chat(ctx, systemPrompt, userPrompt, &result); err != nil {
		return nil, fmt.Errorf("llm resume parse: %w", err)
	}

	// Defensive: ensure non-nil slice for stable JSON output.
	if result.Experiences == nil {
		result.Experiences = []ParsedExperience{}
	}
	return &result, nil
}
