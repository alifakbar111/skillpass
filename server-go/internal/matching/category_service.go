package matching

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type CategoryService struct {
	db  *sql.DB
	bun bun.IDB
}

func NewCategoryService(db *sql.DB, bun bun.IDB) *CategoryService {
	return &CategoryService{db: db, bun: bun}
}

func (s *CategoryService) GetJobCategoryWeights(ctx context.Context, jobPostingID string) (map[string]int, error) {
	jobUUID, err := uuid.Parse(jobPostingID)
	if err != nil {
		return nil, fmt.Errorf("invalid job posting ID: %w", err)
	}

	var results []struct {
		Name   string `bun:"name"`
		Weight int    `bun:"weight"`
	}
	err = s.bun.NewRaw(`
		SELECT sc.name, jcw.weight
		FROM job_category_weights jcw
		INNER JOIN skill_categories sc ON sc.id = jcw.category_id
		WHERE jcw.job_posting_id = ?
	`, jobUUID).Scan(ctx, &results)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	weights := make(map[string]int, len(results))
	for _, r := range results {
		weights[r.Name] = r.Weight
	}
	return weights, nil
}

// GetWeightsForJobs fetches all category weights for a batch of job
// postings in a single query, returning a map keyed by job posting ID.
// Used by the matcher to avoid an N+1 query when scoring many jobs.
func (s *CategoryService) GetWeightsForJobs(ctx context.Context, jobIDs []uuid.UUID) (map[uuid.UUID]map[string]int, error) {
	if len(jobIDs) == 0 {
		return map[uuid.UUID]map[string]int{}, nil
	}
	var rows []struct {
		JobID  string `bun:"job_posting_id"`
		Name   string `bun:"name"`
		Weight int    `bun:"weight"`
	}
	err := s.bun.NewRaw(`
		SELECT jcw.job_posting_id, sc.name, jcw.weight
		FROM job_category_weights jcw
		INNER JOIN skill_categories sc ON sc.id = jcw.category_id
		WHERE jcw.job_posting_id = ANY(?)
	`, bun.In(jobIDs)).Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]map[string]int, len(jobIDs))
	for _, r := range rows {
		jobUUID, err := uuid.Parse(r.JobID)
		if err != nil {
			continue
		}
		bucket, ok := out[jobUUID]
		if !ok {
			bucket = make(map[string]int)
			out[jobUUID] = bucket
		}
		bucket[r.Name] = r.Weight
	}
	return out, nil
}

type SkillCountEntry struct {
	Skill string `json:"skill"`
	Count int    `json:"count"`
}

// categoryNormalize lowercases and trims whitespace for category lookup.
func categoryNormalize(s string) string {
	return strings.TrimSpace(strings.ToLower(s))
}

// AssignSkillsToCategory assigns each skill to its best-matching category.
func AssignSkillsToCategory(skillNames []string) map[string]string {
	skillCategoryMap := map[string]string{
		// Software Engineering
		"go": "Software Engineering", "golang": "Software Engineering",
		"react": "Software Engineering", "react.js": "Software Engineering",
		"python": "Software Engineering", "javascript": "Software Engineering",
		"typescript": "Software Engineering", "rust": "Software Engineering",
		"java": "Software Engineering", "git": "Software Engineering",
		"rest": "Software Engineering", "rest api": "Software Engineering",
		"rest apis": "Software Engineering", "graphql": "Software Engineering",
		"node": "Software Engineering", "node.js": "Software Engineering",
		"nodejs": "Software Engineering", "api": "Software Engineering",
		"microservices": "Software Engineering", "testing": "Software Engineering",
		"ci/cd": "Software Engineering", "grpc": "Software Engineering",
		// Data & Analytics
		"sql": "Data & Analytics", "postgresql": "Data & Analytics",
		"machine learning": "Data & Analytics", "tableau": "Data & Analytics",
		"data engineering": "Data & Analytics", "statistics": "Data & Analytics",
		"data analysis": "Data & Analytics", "data science": "Data & Analytics",
		// Infrastructure & IT
		"aws": "Infrastructure & IT", "docker": "Infrastructure & IT",
		"kubernetes": "Infrastructure & IT", "linux": "Infrastructure & IT",
		"terraform": "Infrastructure & IT", "devops": "Infrastructure & IT",
		// Clinical & Medical
		"patient care": "Clinical & Medical", "triage": "Clinical & Medical",
		"wound care": "Clinical & Medical", "phlebotomy": "Clinical & Medical",
		"anatomy": "Clinical & Medical", "pharmacology": "Clinical & Medical",
		// Management & Leadership
		"team leadership": "Management & Leadership", "agile": "Management & Leadership",
		"project management": "Management & Leadership",
		"stakeholder management": "Management & Leadership",
	}

	result := make(map[string]string, len(skillNames))
	for _, skill := range skillNames {
		lower := categoryNormalize(skill)
		if cat, ok := skillCategoryMap[lower]; ok {
			result[skill] = cat
		} else {
			result[skill] = "Software Engineering"
		}
	}
	return result
}

func (s *CategoryService) ComputeWeightedMatchScore(ctx context.Context, skillCounts []SkillCountEntry, jobPostingID string) (float64, error) {
	weights, err := s.GetJobCategoryWeights(ctx, jobPostingID)
	if err != nil {
		return 0, err
	}
	if weights == nil {
		var total float64
		for _, sc := range skillCounts {
			total += float64(sc.Count)
		}
		return total, nil
	}

	skillNames := make([]string, len(skillCounts))
	for i, sc := range skillCounts {
		skillNames[i] = sc.Skill
	}
	categories := AssignSkillsToCategory(skillNames)

	var total float64
	for _, sc := range skillCounts {
		cat := categories[sc.Skill]
		w := weights[cat]
		if w == 0 {
			w = 1
		}
		total += float64(sc.Count) * float64(w)
	}
	return total, nil
}
