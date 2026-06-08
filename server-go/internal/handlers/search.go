package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type CandidateResult struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	AvatarURL        *string  `json:"avatarUrl"`
	Headline         *string  `json:"headline"`
	About            *string  `json:"about"`
	YearsOfExp       *int     `json:"yearsOfExperience"`
	Slug             string   `json:"slug"`
	Skills           []string `json:"skills"`
}

type SearchHandler struct {
	db *sql.DB
}

func NewSearchHandler(db *sql.DB) *SearchHandler {
	return &SearchHandler{db: db}
}

const (
	searchDefaultLimit = 50
	searchMaxLimit     = 200
	searchMaxPool      = 1000
)

func (h *SearchHandler) SearchCandidates(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	skillsFilter := strings.TrimSpace(c.Query("skills"))
	industry := strings.TrimSpace(c.Query("industry"))

	limit := int64(searchDefaultLimit)
	if v, err := strconv.ParseInt(c.Query("limit"), 10, 64); err == nil && v > 0 && v <= searchMaxLimit {
		limit = v
	}
	offset := int64(0)
	if v, err := strconv.ParseInt(c.Query("offset"), 10, 64); err == nil && v >= 0 {
		offset = v
	}

	var skillList []string
	if skillsFilter != "" {
		for _, s := range strings.Split(skillsFilter, ",") {
			s = strings.TrimSpace(strings.ToLower(s))
			if s != "" {
				skillList = append(skillList, s)
			}
		}
	}

	// Build the search query using raw SQL to avoid go-jet type issues with UUID/enum columns
	var args []interface{}
	argIdx := 1

	whereClauses := []string{"1=1"}
	if q != "" {
		pat := "%" + strings.ToLower(q) + "%"
		whereClauses = append(whereClauses,
			fmt.Sprintf("(LOWER(u.name) LIKE $%d OR LOWER(jp.headline) LIKE $%d OR LOWER(jp.about) LIKE $%d)", argIdx, argIdx, argIdx))
		args = append(args, pat)
		argIdx++
	}

	query := fmt.Sprintf(`
		SELECT jp.id, u.name, u.avatar_url, jp.headline, jp.about, jp.years_of_experience, jp.slug
		FROM jobseeker_profiles jp
		INNER JOIN users u ON jp.user_id = u.id
		WHERE %s
		ORDER BY jp.id ASC
		LIMIT %d OFFSET 0
	`, strings.Join(whereClauses, " AND "), searchMaxPool)

	rows, err := h.db.QueryContext(c.Request.Context(), query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("search failed: %v", err)})
		return
	}
	defer rows.Close()

	type candidateRow struct {
		ID                string
		Name              string
		AvatarURL         *string
		Headline          *string
		About             *string
		YearsOfExperience *int32
		Slug              string
	}
	var profileRows []candidateRow
	for rows.Next() {
		var r candidateRow
		if err := rows.Scan(&r.ID, &r.Name, &r.AvatarURL, &r.Headline, &r.About, &r.YearsOfExperience, &r.Slug); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("scan failed: %v", err)})
			return
		}
		profileRows = append(profileRows, r)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("rows error: %v", err)})
		return
	}

	if len(profileRows) == 0 {
		c.JSON(http.StatusOK, []CandidateResult{})
		return
	}

	// Load experiences for the found profiles
	profileIDs := make([]string, len(profileRows))
	idSet := make(map[string]struct{}, len(profileRows))
	for i, r := range profileRows {
		profileIDs[i] = r.ID
		idSet[r.ID] = struct{}{}
	}

	var expRows []struct {
		ProfileID  string
		Industry   *string
		SkillsRaw  *string
	}
	{
		placeholders := make([]string, len(profileIDs))
		expArgs := make([]interface{}, len(profileIDs))
		for i, pid := range profileIDs {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			expArgs[i] = pid
		}
		expQuery := fmt.Sprintf(
			`SELECT profile_id, industry, skills_used::text FROM job_experiences WHERE profile_id IN (%s)`,
			strings.Join(placeholders, ","),
		)
		eRows, err := h.db.QueryContext(c.Request.Context(), expQuery, expArgs...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to load experiences: %v", err)})
			return
		}
		for eRows.Next() {
			var e struct {
				ProfileID  string
				Industry   *string
				SkillsRaw  *string
			}
			if err := eRows.Scan(&e.ProfileID, &e.Industry, &e.SkillsRaw); err != nil {
				eRows.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("scan experience: %v", err)})
				return
			}
			expRows = append(expRows, e)
		}
		eRows.Close()
	}

	type expAgg struct {
		skillSet map[string]struct{}
		hasInd   bool
	}
	expsByProfile := make(map[string]*expAgg, len(profileRows))
	for _, e := range expRows {
		agg, ok := expsByProfile[e.ProfileID]
		if !ok {
			agg = &expAgg{skillSet: map[string]struct{}{}}
			expsByProfile[e.ProfileID] = agg
		}
		if industry != "" && e.Industry != nil && strings.EqualFold(*e.Industry, industry) {
			agg.hasInd = true
		}
		if e.SkillsRaw != nil {
			// Parse PostgreSQL array format {elem1,elem2} -> individual elements
			raw := *e.SkillsRaw
			if len(raw) >= 2 && raw[0] == '{' && raw[len(raw)-1] == '}' {
				inner := raw[1 : len(raw)-1]
				for _, s := range strings.Split(inner, ",") {
					s = strings.TrimSpace(s)
					if s == "" {
						continue
					}
					agg.skillSet[strings.ToLower(s)] = struct{}{}
				}
			}
		}
	}

	results := make([]CandidateResult, 0, len(profileRows))
	for _, r := range profileRows {
		agg := expsByProfile[r.ID]
		if industry != "" {
			if agg == nil || !agg.hasInd {
				continue
			}
		}
		if len(skillList) > 0 {
			if agg == nil {
				continue
			}
			matched := false
			for _, s := range skillList {
				if _, ok := agg.skillSet[s]; ok {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		skills := make([]string, 0)
		if agg != nil {
			for s := range agg.skillSet {
				skills = append(skills, s)
			}
		}
		results = append(results, CandidateResult{
			ID:         r.ID,
			Name:       r.Name,
			AvatarURL:  r.AvatarURL,
			Headline:   r.Headline,
			About:      r.About,
			YearsOfExp: int32ToIntPtr(r.YearsOfExperience),
			Slug:       r.Slug,
			Skills:     skills,
		})
		if int64(len(results)) >= offset+limit {
			break
		}
	}

	if int64(len(results)) > offset {
		if offset >= int64(len(results)) {
			results = results[len(results):]
		} else {
			results = results[offset:]
		}
		if int64(len(results)) > limit {
			results = results[:limit]
		}
	} else {
		results = results[:0]
	}

	c.JSON(http.StatusOK, results)
}
