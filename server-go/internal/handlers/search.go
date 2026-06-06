package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	. "github.com/go-jet/jet/v2/postgres"

	"skillpass-server-go/internal/gen"
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

	var whereConds []BoolExpression
	if q != "" {
		pat := "%" + strings.ToLower(q) + "%"
		whereConds = append(whereConds, OR(
			LOWER(gen.Users.Name).LIKE(String(pat)),
			LOWER(CAST(gen.JobseekerProfiles.Headline).AS_TEXT()).LIKE(String(pat)),
			LOWER(CAST(gen.JobseekerProfiles.About).AS_TEXT()).LIKE(String(pat)),
		))
	}

	stmt := SELECT(
		gen.JobseekerProfiles.ID,
		gen.Users.Name,
		gen.Users.AvatarURL,
		gen.JobseekerProfiles.Headline,
		gen.JobseekerProfiles.About,
		gen.JobseekerProfiles.YearsOfExperience,
		gen.JobseekerProfiles.Slug,
	).FROM(
		gen.JobseekerProfiles.INNER_JOIN(gen.Users, gen.JobseekerProfiles.UserID.EQ(gen.Users.ID)),
	).WHERE(
		AND(whereConds...),
	).ORDER_BY(
		gen.JobseekerProfiles.ID.ASC(),
	).LIMIT(int64(searchMaxPool)).OFFSET(0)

	type candidateRow struct {
		ID                string
		Name              string
		AvatarURL         *string
		Headline          *string
		About             *string
		YearsOfExperience *int32
		Slug              string
	}

	var rows []candidateRow
	if err := stmt.QueryContext(c.Request.Context(), h.db, &rows); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusOK, []CandidateResult{})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("search failed: %v", err)})
		return
	}

	if len(rows) == 0 {
		c.JSON(http.StatusOK, []CandidateResult{})
		return
	}

	profileIDs := make([]string, 0, len(rows))
	idSet := make(map[string]struct{}, len(rows))
	for _, r := range rows {
		profileIDs = append(profileIDs, r.ID)
		idSet[r.ID] = struct{}{}
	}

	expStmt := SELECT(
		gen.JobExperiences.ProfileID,
		gen.JobExperiences.Industry,
		gen.JobExperiences.SkillsUsed,
	).FROM(
		gen.JobExperiences,
	).WHERE(
		gen.JobExperiences.ProfileID.IN(StringArray(profileIDs...)),
	)

	var exps []struct {
		ProfileID  string
		Industry   *string
		SkillsUsed *string
	}
	if err := expStmt.QueryContext(c.Request.Context(), h.db, &exps); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to load experiences: %v", err)})
		return
	}

	type expAgg struct {
		skillSet map[string]struct{}
		hasInd   bool
	}
	expsByProfile := make(map[string]*expAgg, len(rows))
	for _, e := range exps {
		agg, ok := expsByProfile[e.ProfileID]
		if !ok {
			agg = &expAgg{skillSet: map[string]struct{}{}}
			expsByProfile[e.ProfileID] = agg
		}
		if industry != "" && e.Industry != nil && strings.EqualFold(*e.Industry, industry) {
			agg.hasInd = true
		}
		if e.SkillsUsed != nil {
			for _, s := range strings.Split(*e.SkillsUsed, ",") {
				s = strings.TrimSpace(s)
				if s == "" {
					continue
				}
				agg.skillSet[strings.ToLower(s)] = struct{}{}
			}
		}
	}

	results := make([]CandidateResult, 0, len(rows))
	for _, r := range rows {
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
