package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	. "github.com/go-jet/jet/v2/postgres"
	"database/sql"

	"skillpass-server-go/.gen/skillpass/public/model"
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

type searchExp struct {
	Title        string
	Organization string
	Industry     *string
	SkillsUsed   []string
}

func (h *SearchHandler) SearchCandidates(c *gin.Context) {
	q := c.Query("q")
	skills := c.Query("skills")
	industry := c.Query("industry")

	profStmt := SELECT(
		gen.JobseekerProfiles.ID, gen.JobseekerProfiles.UserID, gen.JobseekerProfiles.Headline,
		gen.JobseekerProfiles.About, gen.JobseekerProfiles.YearsOfExperience, gen.JobseekerProfiles.Slug,
	).FROM(
		gen.JobseekerProfiles,
	)

	var profiles []model.JobseekerProfiles
	err := profStmt.QueryContext(c.Request.Context(), h.db, &profiles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query profiles"})
		return
	}

	results := make([]CandidateResult, 0)

	for _, p := range profiles {
		var user model.Users
		userStmt := SELECT(
			gen.Users.Name, gen.Users.AvatarURL,
		).FROM(
			gen.Users,
		).WHERE(
			gen.Users.ID.EQ(String(p.UserID.String())),
		)
		err := userStmt.QueryContext(c.Request.Context(), h.db, &user)
		if err != nil {
			continue
		}

		var exps []model.JobExperiences
		expStmt := SELECT(
			gen.JobExperiences.Title, gen.JobExperiences.Organization,
			gen.JobExperiences.Industry, gen.JobExperiences.SkillsUsed,
		).FROM(
			gen.JobExperiences,
		).WHERE(
			gen.JobExperiences.ProfileID.EQ(String(p.ID.String())),
		)
		err = expStmt.QueryContext(c.Request.Context(), h.db, &exps)
		if err != nil {
			continue
		}

		expList := make([]searchExp, len(exps))
		for i, e := range exps {
			expList[i].Title = e.Title
			expList[i].Organization = e.Organization
			expList[i].Industry = e.Industry
			if e.SkillsUsed != nil {
				expList[i].SkillsUsed = []string(*e.SkillsUsed)
			}
		}

		// Filter by search query
		if q != "" {
			ql := strings.ToLower(q)
			matchesName := strings.Contains(strings.ToLower(user.Name), ql)
			matchesHeadline := p.Headline != nil && strings.Contains(strings.ToLower(*p.Headline), ql)
			matchesAbout := p.About != nil && strings.Contains(strings.ToLower(*p.About), ql)
			matchesExp := false
			for _, e := range expList {
				if strings.Contains(strings.ToLower(e.Title), ql) ||
					strings.Contains(strings.ToLower(e.Organization), ql) {
					matchesExp = true
					break
				}
				for _, s := range e.SkillsUsed {
					if strings.Contains(strings.ToLower(s), ql) {
						matchesExp = true
						break
					}
				}
			}
			if !matchesName && !matchesHeadline && !matchesAbout && !matchesExp {
				continue
			}
		}

		// Filter by skills
		if skills != "" {
			skillList := strings.Split(skills, ",")
			for i := range skillList {
				skillList[i] = strings.TrimSpace(strings.ToLower(skillList[i]))
			}
			hasSkill := false
			for _, e := range expList {
				for _, s := range e.SkillsUsed {
					for _, sk := range skillList {
						if strings.EqualFold(s, sk) {
							hasSkill = true
							break
						}
					}
					if hasSkill {
						break
					}
				}
				if hasSkill {
					break
				}
			}
			if !hasSkill {
				continue
			}
		}

		// Filter by industry
		if industry != "" {
			hasIndustry := false
			for _, e := range expList {
				if e.Industry != nil && strings.EqualFold(*e.Industry, industry) {
					hasIndustry = true
					break
				}
			}
			if !hasIndustry {
				continue
			}
		}

		// Collect unique skills
		skillSet := make(map[string]struct{})
		for _, e := range expList {
			for _, s := range e.SkillsUsed {
				skillSet[s] = struct{}{}
			}
		}
		uniqueSkills := make([]string, 0, len(skillSet))
		for s := range skillSet {
			uniqueSkills = append(uniqueSkills, s)
		}

		results = append(results, CandidateResult{
			ID:         p.ID.String(),
			Name:       user.Name,
			AvatarURL:  user.AvatarURL,
			Headline:   p.Headline,
			About:      p.About,
			YearsOfExp: int32ToIntPtr(p.YearsOfExperience),
			Slug:       p.Slug,
			Skills:     uniqueSkills,
		})
	}

	c.JSON(http.StatusOK, results)
}
