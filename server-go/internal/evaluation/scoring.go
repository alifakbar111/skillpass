package evaluation

// RoleWeight represents the seniority bucket for a skill.
type RoleWeight string

const (
	RoleEntry   RoleWeight = "entry"
	RoleSkilled RoleWeight = "skilled"
	RoleSenior  RoleWeight = "senior"
	RoleExpert  RoleWeight = "expert"
)

var roleWeightPoints = map[RoleWeight]int{
	RoleEntry:   10,
	RoleSkilled: 35,
	RoleSenior:  65,
	RoleExpert:  100,
}

// EducationLevel represents the highest education level for a skill.
type EducationLevel string

const (
	EduNone     EducationLevel = "none"
	EduHS       EducationLevel = "hs"
	EduDiploma  EducationLevel = "diploma"
	EduBachelor EducationLevel = "bachelor"
	EduMaster   EducationLevel = "master"
	EduPhD      EducationLevel = "phd"
)

var eduPoints = map[EducationLevel]int{
	EduNone:     0,
	EduHS:       5,
	EduDiploma:  15,
	EduBachelor: 30,
	EduMaster:   45,
	EduPhD:      65,
}

// SkillFacts are the LLM-extracted structured facts per skill.
type SkillFacts struct {
	Skill             string         `json:"skill"`
	TotalYears        float64        `json:"totalYears"`
	NumRoles          int            `json:"numRoles"`
	RoleWeight        RoleWeight     `json:"roleWeight"`
	EducationLevel    EducationLevel `json:"educationLevel"`
	NumCertifications int            `json:"numCertifications"`
	NumLicenses       int            `json:"numLicenses"`
	NumProjects       int            `json:"numProjects"`
	NumOrganizations  int            `json:"numOrganizations"`
	HasURL            bool           `json:"hasUrl"`
}

// SkillCountResult holds the computed Count and its breakdown for a skill.
type SkillCountResult struct {
	Skill               string `json:"skill"`
	Count               int    `json:"count"`
	YearsPoints         int    `json:"yearsPoints"`
	RolesPoints         int    `json:"rolesPoints"`
	RoleWeightPoints    int    `json:"roleWeightPoints"`
	EducationPoints     int    `json:"educationPoints"`
	CertificationPoints int    `json:"certificationPoints"`
	ProjectPoints       int    `json:"projectPoints"`
	DiversityPoints     int    `json:"diversityPoints"`
	URLPoints           int    `json:"urlPoints"`
}

// ComputeSkillCount calculates the deterministic Count for one skill.
func ComputeSkillCount(facts SkillFacts) SkillCountResult {
	yearsPoints := int(facts.TotalYears*10 + 0.5)
	rolesPoints := facts.NumRoles * 15
	rwPoints := roleWeightPoints[facts.RoleWeight]
	if rwPoints == 0 {
		rwPoints = 10
	}
	ep := eduPoints[facts.EducationLevel]
	certPoints := facts.NumCertifications*10 + facts.NumLicenses*20
	projectPoints := facts.NumProjects * 10
	orgPoints := 0
	if facts.NumOrganizations > 1 {
		orgPoints = (facts.NumOrganizations - 1) * 5
	}
	urlPoints := 0
	if facts.HasURL {
		urlPoints = 10
	}

	total := yearsPoints + rolesPoints + rwPoints + ep + certPoints + projectPoints + orgPoints + urlPoints

	return SkillCountResult{
		Skill:               facts.Skill,
		Count:               total,
		YearsPoints:         yearsPoints,
		RolesPoints:         rolesPoints,
		RoleWeightPoints:    rwPoints,
		EducationPoints:     ep,
		CertificationPoints: certPoints,
		ProjectPoints:       projectPoints,
		DiversityPoints:     orgPoints,
		URLPoints:           urlPoints,
	}
}

// ComputeTotalCount calculates the overall total Count from per-skill results.
func ComputeTotalCount(skills []SkillCountResult) int {
	sum := 0
	for _, s := range skills {
		sum += s.Count
	}
	if len(skills) > 1 {
		sum += (len(skills) - 1) * 10
	}
	return sum
}
