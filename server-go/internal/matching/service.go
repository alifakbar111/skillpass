package matching

import (
	"context"
	"database/sql"
	"errors"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"skillpass-server-go/.gen/skillpass/public/model"
	"skillpass-server-go/internal/gen"
	"skillpass-server-go/internal/lib"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

type JobMatch struct {
	JobPostingID    string  `json:"jobPostingId"`
	Title           string  `json:"title"`
	CompanyName     string  `json:"companyName"`
	Industry        string  `json:"industry"`
	Location        *string `json:"location"`
	SalaryRange     *string `json:"salaryRange"`
	ExperienceLevel *string `json:"experienceLevel"`
	MatchScore      float64 `json:"matchScore"`
	MatchReason     string  `json:"matchReason"`
} //@name JobMatch

type CandidateMatch struct {
	ProfileID    string   `json:"profileId"`
	Name         string   `json:"name"`
	Headline     *string  `json:"headline"`
	OverallScore int32    `json:"overallScore"`
	TopSkills    []string `json:"topSkills"`
	MatchScore   float64  `json:"matchScore"`
	MatchReason  string   `json:"matchReason"`
} //@name CandidateMatch

type skillScoreData struct {
	Skill    string  `json:"skill"`
	Category string  `json:"category"`
	Score    float64 `json:"score"`
}

type jobMatchRow struct {
	ID              uuid.UUID
	Title           string
	Industry        string
	RequiredSkills  pq.StringArray
	Location        *string
	SalaryRange     *string
	ExperienceLevel *string
	CompanyName     string
}

func (s *Service) MatchJobs(ctx context.Context, profileID string) ([]JobMatch, error) {
	eval, err := s.getLatestEvaluation(ctx, profileID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get evaluation: %w", err)
	}

	skillNames := extractSkillNames(eval)
	if len(skillNames) == 0 {
		return nil, nil
	}

	// Build candidate skill map once (optimization)
	candidateMap := make(map[string]bool, len(skillNames))
	for _, s := range skillNames {
		candidateMap[strings.ToLower(s)] = true
	}

	// Raw SQL: go-jet's qrm does not reliably scan joined columns and text[]
	// arrays into ad-hoc slice destinations (rows silently come back empty).
	jobRows, err := s.db.QueryContext(ctx, `
		SELECT j.id, j.title, j.industry, COALESCE(j.required_skills, '{}'),
		       j.location, j.salary_range, j.experience_level::text, c.company_name
		FROM job_postings j
		JOIN companies c ON c.id = j.company_id
		WHERE j.status = 'open'
		ORDER BY j.created_at DESC
		LIMIT 200`)
	if err != nil {
		return nil, fmt.Errorf("query open jobs: %w", err)
	}
	defer jobRows.Close()

	var rows []jobMatchRow
	for jobRows.Next() {
		var r jobMatchRow
		if err := jobRows.Scan(&r.ID, &r.Title, &r.Industry, &r.RequiredSkills,
			&r.Location, &r.SalaryRange, &r.ExperienceLevel, &r.CompanyName); err != nil {
			return nil, fmt.Errorf("scan job row: %w", err)
		}
		rows = append(rows, r)
	}
	if err := jobRows.Err(); err != nil {
		return nil, err
	}

	type scored struct {
		row   jobMatchRow
		score float64
	}
	var scoredJobs []scored

	for _, row := range rows {
		score := computeMatchScoreWithMap(candidateMap, []string(row.RequiredSkills))
		if score > 0 {
			scoredJobs = append(scoredJobs, scored{row: row, score: score})
		}
	}

	sort.Slice(scoredJobs, func(i, j int) bool {
		return scoredJobs[i].score > scoredJobs[j].score
	})

	if len(scoredJobs) > 20 {
		scoredJobs = scoredJobs[:20]
	}

	results := make([]JobMatch, len(scoredJobs))
	for i, sj := range scoredJobs {
		results[i] = JobMatch{
			JobPostingID:    sj.row.ID.String(),
			Title:           sj.row.Title,
			CompanyName:     sj.row.CompanyName,
			Industry:        sj.row.Industry,
			Location:        sj.row.Location,
			SalaryRange:     sj.row.SalaryRange,
			ExperienceLevel: sj.row.ExperienceLevel,
			MatchScore:      sj.score,
			MatchReason:     computeReason(sj.score),
		}
	}

	return results, nil
}

func (s *Service) MatchCandidates(ctx context.Context, jobPostingID string) ([]CandidateMatch, error) {
	var jobSkillsArr pq.StringArray
	err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(required_skills, '{}') FROM job_postings WHERE id = $1`,
		jobPostingID,
	).Scan(&jobSkillsArr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query job: %w", err)
	}

	jobSkills := []string(jobSkillsArr)
	if len(jobSkills) == 0 {
		return nil, nil
	}

	stmt := SELECT(
		gen.AiEvaluations.ProfileID,
		gen.AiEvaluations.OverallScore,
		gen.AiEvaluations.SkillScores,
		gen.JobseekerProfiles.Headline,
		gen.Users.Name,
	).FROM(
		gen.AiEvaluations.
			INNER_JOIN(gen.JobseekerProfiles, gen.JobseekerProfiles.ID.EQ(gen.AiEvaluations.ProfileID)).
			INNER_JOIN(gen.Users, gen.Users.ID.EQ(gen.JobseekerProfiles.UserID)),
	).ORDER_BY(
		gen.AiEvaluations.CreatedAt.DESC(),
	).LIMIT(200)

	var rows []struct {
		ProfileID    uuid.UUID `alias:"ai_evaluations.profile_id"`
		OverallScore int32     `alias:"ai_evaluations.overall_score"`
		SkillScores  string    `alias:"ai_evaluations.skill_scores"`
		Headline     *string   `alias:"jobseeker_profiles.headline"`
		Name         string    `alias:"users.name"`
	}
	if err := stmt.QueryContext(ctx, s.db, &rows); err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	type candidateEval struct {
		ProfileID    string
		Name         string
		Headline     *string
		OverallScore int32
		SkillScores  []skillScoreData
	}
	var candidates []candidateEval

	for _, row := range rows {
		pid := row.ProfileID.String()
		if seen[pid] {
			continue
		}
		seen[pid] = true

		var skills []skillScoreData
		if err := json.Unmarshal([]byte(row.SkillScores), &skills); err != nil {
			slog.Warn("failed to unmarshal skill scores for candidate", "profileID", pid, "error", err)
			continue
		}

		candidates = append(candidates, candidateEval{
			ProfileID:    pid,
			Name:         row.Name,
			Headline:     row.Headline,
			OverallScore: row.OverallScore,
			SkillScores:  skills,
		})
	}

	type scoredCandidate struct {
		candidateEval
		score float64
	}
	var scoredCandidates []scoredCandidate

	for _, c := range candidates {
		candidateSkills := make([]string, len(c.SkillScores))
		for i, ss := range c.SkillScores {
			candidateSkills[i] = strings.ToLower(ss.Skill)
		}
		// Source = candidate skills (what % of job requirements do they meet)
		score := computeMatchScore(candidateSkills, jobSkills)
		if score > 0 {
			scoredCandidates = append(scoredCandidates, scoredCandidate{candidateEval: c, score: score})
		}
	}

	sort.Slice(scoredCandidates, func(i, j int) bool {
		return scoredCandidates[i].score > scoredCandidates[j].score
	})

	if len(scoredCandidates) > 20 {
		scoredCandidates = scoredCandidates[:20]
	}

	results := make([]CandidateMatch, len(scoredCandidates))
	for i, sc := range scoredCandidates {
		topSkills := make([]string, 0, 5)
		for _, ss := range sc.SkillScores {
			topSkills = append(topSkills, ss.Skill)
			if len(topSkills) >= 5 {
				break
			}
		}
		results[i] = CandidateMatch{
			ProfileID:    sc.ProfileID,
			Name:         sc.Name,
			Headline:     sc.Headline,
			OverallScore: sc.OverallScore,
			TopSkills:    topSkills,
			MatchScore:   sc.score,
			MatchReason:  computeReason(sc.score),
		}
	}

	return results, nil
}

// SkillsGap compares a jobseeker's evaluated skills against a job's required skills.
type SkillsGap struct {
	JobPostingID  string   `json:"jobPostingId"`
	JobTitle      string   `json:"jobTitle"`
	MatchedSkills []string `json:"matchedSkills"`
	MissingSkills []string `json:"missingSkills"`
	MatchPercent  float64  `json:"matchPercent"`
	HasEvaluation bool     `json:"hasEvaluation"`
}

// ComputeSkillsGap returns which of the job's required skills the candidate has
// (per their latest AI evaluation) and which are missing.
func (s *Service) ComputeSkillsGap(ctx context.Context, profileID, jobPostingID string) (*SkillsGap, error) {
	var jobTitle string
	var jobSkillsArr pq.StringArray
	err := s.db.QueryRowContext(ctx,
		`SELECT title, COALESCE(required_skills, '{}') FROM job_postings WHERE id = $1`,
		jobPostingID,
	).Scan(&jobTitle, &jobSkillsArr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("query job: %w", err)
	}

	gap := &SkillsGap{
		JobPostingID:  jobPostingID,
		JobTitle:      jobTitle,
		MatchedSkills: []string{},
		MissingSkills: []string{},
	}

	jobSkills := []string(jobSkillsArr)
	if len(jobSkills) == 0 {
		gap.MatchPercent = 100
		gap.HasEvaluation = true
		return gap, nil
	}

	candidateSet := map[string]bool{}
	eval, err := s.getLatestEvaluation(ctx, profileID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("get evaluation: %w", err)
		}
		// No evaluation yet — everything counts as missing.
		gap.MissingSkills = jobSkills
		return gap, nil
	}
	gap.HasEvaluation = true
	for _, name := range extractSkillNames(eval) {
		candidateSet[name] = true
	}

	for _, skill := range jobSkills {
		if candidateSet[strings.ToLower(skill)] {
			gap.MatchedSkills = append(gap.MatchedSkills, skill)
		} else {
			gap.MissingSkills = append(gap.MissingSkills, skill)
		}
	}
	gap.MatchPercent = float64(len(gap.MatchedSkills)) / float64(len(jobSkills)) * 100

	return gap, nil
}

// IsBlindMode reports whether the company has blind screening enabled.
func (s *Service) IsBlindMode(ctx context.Context, companyID string) bool {
	var blind bool
	if err := s.db.QueryRowContext(ctx,
		`SELECT blind_mode FROM companies WHERE id = $1`, companyID,
	).Scan(&blind); err != nil {
		return false
	}
	return blind
}

func (s *Service) getLatestEvaluation(ctx context.Context, profileID string) (*model.AiEvaluations, error) {
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
	).LIMIT(1)

	var evals []model.AiEvaluations
	err = stmt.QueryContext(ctx, s.db, &evals)
	if err != nil {
		return nil, err
	}
	if len(evals) == 0 {
		return nil, sql.ErrNoRows
	}
	return &evals[0], nil
}

func extractSkillNames(eval *model.AiEvaluations) []string {
	var scores []skillScoreData
	if err := json.Unmarshal([]byte(eval.SkillScores), &scores); err != nil {
		slog.Warn("failed to unmarshal skill scores for latest evaluation", "error", err)
		return nil
	}

	skillSet := map[string]bool{}
	for _, s := range scores {
		if s.Score >= 50 {
			skillSet[strings.ToLower(s.Skill)] = true
		}
	}

	skills := make([]string, 0, len(skillSet))
	for s := range skillSet {
		skills = append(skills, s)
	}
	return skills
}

func computeMatchScore(sourceSkills, targetSkills []string) float64 {
	if len(sourceSkills) == 0 || len(targetSkills) == 0 {
		return 0
	}

	sourceMap := make(map[string]bool, len(sourceSkills))
	for _, s := range sourceSkills {
		sourceMap[strings.ToLower(s)] = true
	}
	return computeMatchScoreWithMap(sourceMap, targetSkills)
}

// computeMatchScoreWithMap uses a pre-built source map for efficiency.
// sourceMap is a map of lowercased skill names from the source (candidate).
// targetSkills are skills from the job posting.
func computeMatchScoreWithMap(sourceMap map[string]bool, targetSkills []string) float64 {
	if len(targetSkills) == 0 {
		return 0
	}

	matches := 0
	for _, t := range targetSkills {
		if sourceMap[strings.ToLower(t)] {
			matches++
		}
	}

	overlap := float64(matches) / float64(len(targetSkills))
	relevance := float64(matches) / float64(len(sourceMap))

	score := (overlap*0.6 + relevance*0.4) * 100
	return score
}

func computeReason(score float64) string {
	switch {
	case score >= 80:
		return "Strong match — skills align well"
	case score >= 60:
		return "Good match — most skills overlap"
	case score >= 40:
		return "Moderate match — some skills align"
	default:
		return "Partial match — consider reviewing details"
	}
}
