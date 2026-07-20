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

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/uptrace/bun"

	"skillpass-server-go/internal/lib"
	"skillpass-server-go/internal/models"
)

type Service struct {
	db              *sql.DB
	bun             bun.IDB
	categoryService *CategoryService
}

func (s *Service) SetCategoryService(cs *CategoryService) {
	s.categoryService = cs
}

func NewService(db *sql.DB, bun bun.IDB) *Service {
	return &Service{db: db, bun: bun}
}

type JobMatch struct {
	JobPostingID    string  `json:"jobPostingId"`
	Title           string  `json:"title"`
	CompanyName     string  `json:"companyName"`
	Industry        string  `json:"industry"`
	Location        *string `json:"location,omitempty"`
	SalaryRange     *string `json:"salaryRange,omitempty"`
	ExperienceLevel *string `json:"experienceLevel,omitempty"`
	MatchScore      float64 `json:"matchScore"`
	MatchReason     string  `json:"matchReason"`
} //@name JobMatch

type CandidateMatch struct {
	ProfileID    string   `json:"profileId"`
	Name         string   `json:"name"`
	Username     string   `json:"username"`
	Headline     *string  `json:"headline,omitempty"`
	OverallScore int32    `json:"overallScore"`
	TopSkills    []string `json:"topSkills,omitempty"`
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

type scored struct {
	row   jobMatchRow
	score float64
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

	// Build candidate skill map once (optimization).
	// skillNames are already normalized, but re-normalize for safety.
	candidateMap := make(map[string]bool, len(skillNames))
	for _, s := range skillNames {
		candidateMap[normalizeSkill(s)] = true
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

	scoredJobs := s.scoreJobsWithBatchWeights(ctx, rows, candidateMap)

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

// scoreJobsWithBatchWeights scores all jobs in a single pass, fetching
// per-job category weights in one query instead of one-per-job.
func (s *Service) scoreJobsWithBatchWeights(ctx context.Context, rows []jobMatchRow, candidateMap map[string]bool) []scored {
	if len(rows) == 0 {
		return nil
	}
	// Pre-fetch weights for all candidate jobs in one query (avoids N+1).
	var weightsByJob map[uuid.UUID]map[string]int
	if s.categoryService != nil {
		ids := make([]uuid.UUID, len(rows))
		for i, r := range rows {
			ids[i] = r.ID
		}
		if w, err := s.categoryService.GetWeightsForJobs(ctx, ids); err == nil {
			weightsByJob = w
		} else {
			slog.Warn("batch weight fetch failed; falling back to per-row scoring", "error", err)
		}
	}

	var scoredJobs []scored
	for _, row := range rows {
		var score float64
		if s.categoryService != nil {
			weights := weightsByJob[row.ID]
			if weights == nil {
				weights = map[string]int{} // no weights configured → unweighted sum
			}
			score = computeWeightedScoreWithMap(candidateMap, []string(row.RequiredSkills), weights)
		} else {
			score = computeMatchScoreWithMap(candidateMap, []string(row.RequiredSkills))
		}
		if score > 0 {
			scoredJobs = append(scoredJobs, scored{row: row, score: score})
		}
	}
	return scoredJobs
}

// computeWeightedScoreWithMap mirrors computeJobMatchScore but takes a
// pre-fetched weights map instead of hitting the DB per row.
func computeWeightedScoreWithMap(candidateMap map[string]bool, targetSkills []string, weights map[string]int) float64 {
	var matched []SkillCountEntry
	for _, t := range targetSkills {
		if candidateMap[normalizeSkill(t)] {
			matched = append(matched, SkillCountEntry{Skill: t, Count: 1})
		}
	}
	if len(matched) == 0 {
		return 0
	}
	if len(weights) == 0 {
		var total float64
		for _, sc := range matched {
			total += float64(sc.Count)
		}
		return total
	}
	skillNames := make([]string, len(matched))
	for i, sc := range matched {
		skillNames[i] = sc.Skill
	}
	categories := AssignSkillsToCategory(skillNames)
	var total float64
	for _, sc := range matched {
		cat := categories[sc.Skill]
		w := weights[cat]
		if w == 0 {
			w = 1
		}
		total += float64(sc.Count) * float64(w)
	}
	return total
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

	var rows []struct {
		ProfileID    uuid.UUID `bun:"profile_id"`
		OverallScore int32     `bun:"overall_score"`
		SkillScores  string    `bun:"skill_scores"`
		Headline     *string   `bun:"headline"`
		Name         string    `bun:"name"`
		Username     string    `bun:"username"`
	}
	if err := s.bun.NewRaw(`
		SELECT ae.profile_id, ae.overall_score, ae.skill_scores,
		       jp.headline, u.name, u.username
		FROM ai_evaluations ae
		INNER JOIN jobseeker_profiles jp ON jp.id = ae.profile_id
		INNER JOIN users u ON u.id = jp.user_id
		ORDER BY ae.created_at DESC
		LIMIT 200
	`).Scan(ctx, &rows); err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	type candidateEval struct {
		ProfileID    string
		Name         string
		Username     string
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
			Username:     row.Username,
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
		candidateSkills := make([]string, 0, len(c.SkillScores))
		for _, ss := range c.SkillScores {
			candidateSkills = append(candidateSkills, strings.ToLower(ss.Skill))
		}
		// Source = candidate skills (what % of job requirements do they meet)
		score := s.computeCandidateMatchScore(ctx, candidateSkills, jobSkills, jobPostingID)
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
			Username:     sc.Username,
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
	MatchedSkills []string `json:"matchedSkills,omitempty"`
	MissingSkills []string `json:"missingSkills,omitempty"`
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

func (s *Service) getLatestEvaluation(ctx context.Context, profileID string) (*models.Evaluation, error) {
	profileUUID, err := lib.ParseUUID(profileID)
	if err != nil {
		return nil, fmt.Errorf("invalid profile ID: %w", err)
	}

	var eval models.Evaluation
	err = s.bun.NewSelect().
		Model(&eval).
		Where("profile_id = ?", profileUUID).
		Order("created_at DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &eval, nil
}

// normalizeSkill strips punctuation and whitespace for fuzzy skill name matching.
// e.g. "Next.js" → "nextjs", "CI/CD" → "cicd", "REST API" → "restapi"
func normalizeSkill(name string) string {
	n := strings.ToLower(name)
	n = strings.NewReplacer(".", "", "-", "", "/", "", " ", "", "'", "").Replace(n)
	return n
}

func extractSkillNames(eval *models.Evaluation) []string {
	var scores []skillScoreData
	if err := json.Unmarshal([]byte(eval.SkillScores), &scores); err != nil {
		slog.Warn("failed to unmarshal skill scores for latest evaluation", "error", err)
		return nil
	}

	// Find the maximum score to determine the scale the LLM used.
	// The LLM scores individual skills on a variable scale (observed 1-20 per skill,
	// not a fixed 0-100). A relative threshold adapts to whatever scale it uses.
	maxScore := 0.0
	for _, s := range scores {
		if s.Score > maxScore {
			maxScore = s.Score
		}
	}
	threshold := maxScore * 0.5

	skillSet := map[string]bool{}
	for _, s := range scores {
		if s.Score >= threshold {
			skillSet[normalizeSkill(s.Skill)] = true
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
		sourceMap[normalizeSkill(s)] = true
	}
	return computeMatchScoreWithMap(sourceMap, targetSkills)
}

// computeMatchScoreWithMap uses a pre-built source map for efficiency.
// sourceMap is a map of normalized skill names from the source (candidate).
// targetSkills are skills from the job posting.
func computeMatchScoreWithMap(sourceMap map[string]bool, targetSkills []string) float64 {
	if len(targetSkills) == 0 {
		return 0
	}

	matches := 0
	for _, t := range targetSkills {
		if sourceMap[normalizeSkill(t)] {
			matches++
		}
	}

	overlap := float64(matches) / float64(len(targetSkills))
	relevance := float64(matches) / float64(len(sourceMap))

	score := (overlap*0.6 + relevance*0.4) * 100
	return score
}

// computeJobMatchScore uses weighted scoring when categoryService is available.
func (s *Service) computeJobMatchScore(ctx context.Context, candidateMap map[string]bool, targetSkills []string, jobPostingID string) float64 {
	if s.categoryService != nil {
		var matched []SkillCountEntry
		for _, t := range targetSkills {
			if candidateMap[normalizeSkill(t)] {
				matched = append(matched, SkillCountEntry{Skill: t, Count: 1})
			}
		}
		if len(matched) > 0 {
			score, err := s.categoryService.ComputeWeightedMatchScore(ctx, matched, jobPostingID)
			if err == nil {
				return score
			}
		}
	}
	return computeMatchScoreWithMap(candidateMap, targetSkills)
}

// computeCandidateMatchScore uses weighted scoring when categoryService is available.
func (s *Service) computeCandidateMatchScore(ctx context.Context, candidateSkills, jobSkills []string, jobPostingID string) float64 {
	if s.categoryService != nil {
		candidateSet := make(map[string]bool, len(candidateSkills))
		for _, sk := range candidateSkills {
			candidateSet[normalizeSkill(sk)] = true
		}
		var matched []SkillCountEntry
		for _, js := range jobSkills {
			if candidateSet[normalizeSkill(js)] {
				matched = append(matched, SkillCountEntry{Skill: js, Count: 1})
			}
		}
		if len(matched) > 0 {
			score, err := s.categoryService.ComputeWeightedMatchScore(ctx, matched, jobPostingID)
			if err == nil {
				return score
			}
		}
	}
	return computeMatchScore(candidateSkills, jobSkills)
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
