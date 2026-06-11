package handlers

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"

	"skillpass-server-go/internal/email"
)

// ogPageTmpl renders a minimal HTML document whose only job is to give link
// crawlers (Slack, LinkedIn, WhatsApp, X) per-profile Open Graph metadata.
// Human visitors are redirected to the SPA passport route immediately.
var ogPageTmpl = template.Must(template.New("og").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>{{.Title}}</title>
<meta name="description" content="{{.Description}}">
<meta property="og:type" content="profile">
<meta property="og:site_name" content="SkillPass">
<meta property="og:title" content="{{.Title}}">
<meta property="og:description" content="{{.Description}}">
<meta property="og:url" content="{{.URL}}">
{{if .Image}}<meta property="og:image" content="{{.Image}}">{{end}}
<meta name="twitter:card" content="summary">
<meta name="twitter:title" content="{{.Title}}">
<meta name="twitter:description" content="{{.Description}}">
<meta http-equiv="refresh" content="0;url={{.URL}}">
</head>
<body>
<p>Redirecting to <a href="{{.URL}}">{{.Title}}</a>…</p>
<script>window.location.replace('{{.URL}}');</script>
</body>
</html>`))

// GetOGPage	godoc
// @Summary		Passport link-preview page
// @Description	Server-rendered HTML with per-profile Open Graph tags for crawlers; redirects humans to the SPA passport. Does not increment the view counter.
// @Tags		passport
// @Produce		html
// @Param		username path string true "Profile username/slug"
// @Success		200 {string} string
// @Failure		404 {object} map[string]string
// @Router		/p/{username} [get]
func (h *PassportHandler) GetOGPage(c *gin.Context) {
	username := c.Param("username")

	var name string
	var headline, about, avatarURL *string
	err := h.db.QueryRowContext(c.Request.Context(),
		`SELECT u.name, jp.headline, jp.about, u.avatar_url
		 FROM jobseeker_profiles jp
		 JOIN users u ON u.id = jp.user_id
		 WHERE jp.slug = $1`,
		username,
	).Scan(&name, &headline, &about, &avatarURL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}

	title := name + " — SkillPass"
	if headline != nil && *headline != "" {
		title = name + " · " + *headline + " — SkillPass"
	}
	description := "View this verified skill passport on SkillPass."
	if about != nil && *about != "" {
		description = *about
		if len(description) > 200 {
			description = description[:197] + "..."
		}
	}

	base := email.AppBaseURL()
	image := ""
	if avatarURL != nil && *avatarURL != "" {
		image = *avatarURL
		// Local upload paths need to be absolute for crawlers.
		if image[0] == '/' {
			image = base + image
		}
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(http.StatusOK)
	_ = ogPageTmpl.Execute(c.Writer, map[string]string{
		"Title":       title,
		"Description": description,
		"URL":         base + "/profiles/" + username,
		"Image":       image,
	})
}
