package email

import (
	"fmt"
	"html/template"
	"strings"
)

// Message is the rendered form of a transactional email.
type Message struct {
	Subject string
	HTML    string
	Text    string
}

type templateData struct {
	Heading  string
	Lines    []string
	CTAURL   string
	CTALabel string
}

// layoutTmpl is a deliberately simple inline-styled layout that renders
// consistently across mail clients. Values are escaped by html/template.
var layoutTmpl = template.Must(template.New("layout").Parse(`<!DOCTYPE html>
<html>
<body style="margin:0;padding:0;background:#f4f4f5;font-family:Arial,Helvetica,sans-serif;">
  <div style="max-width:560px;margin:24px auto;background:#ffffff;border-radius:8px;padding:32px;">
    <p style="font-size:18px;font-weight:bold;color:#4f46e5;margin:0 0 16px;">SkillPass</p>
    <h2 style="font-size:20px;color:#18181b;margin:0 0 16px;">{{.Heading}}</h2>
    {{range .Lines}}<p style="font-size:14px;color:#3f3f46;line-height:1.6;margin:0 0 12px;">{{.}}</p>{{end}}
    {{if .CTAURL}}<p style="margin:24px 0;"><a href="{{.CTAURL}}" style="background:#4f46e5;color:#ffffff;text-decoration:none;padding:10px 20px;border-radius:6px;font-size:14px;">{{.CTALabel}}</a></p>
    <p style="font-size:12px;color:#a1a1aa;margin:0 0 8px;">Or copy this link: {{.CTAURL}}</p>{{end}}
    <p style="font-size:12px;color:#a1a1aa;margin:24px 0 0;">You received this email because you have a SkillPass account.</p>
  </div>
</body>
</html>`))

func render(subject string, data templateData) Message {
	var html strings.Builder
	if err := layoutTmpl.Execute(&html, data); err != nil {
		// Template is static and tested; fall back to plain text on the
		// impossible path rather than failing the send.
		html.Reset()
		html.WriteString(data.Heading)
	}

	var text strings.Builder
	text.WriteString(data.Heading + "\n\n")
	for _, line := range data.Lines {
		text.WriteString(line + "\n")
	}
	if data.CTAURL != "" {
		text.WriteString("\n" + data.CTALabel + ": " + data.CTAURL + "\n")
	}

	return Message{Subject: subject, HTML: html.String(), Text: text.String()}
}

// VerificationEmail doubles as the welcome email: sent right after
// registration with the email-verification link.
func VerificationEmail(name, verifyURL string) Message {
	return render("Welcome to SkillPass — verify your email", templateData{
		Heading: fmt.Sprintf("Welcome to SkillPass, %s!", name),
		Lines: []string{
			"You're one step away from your skill passport. Please confirm your email address so companies and candidates can trust your account.",
			"This link expires in 24 hours.",
		},
		CTAURL:   verifyURL,
		CTALabel: "Verify my email",
	})
}

// PasswordResetEmail carries the reset link. Sent only to existing accounts.
func PasswordResetEmail(name, resetURL string) Message {
	return render("Reset your SkillPass password", templateData{
		Heading: fmt.Sprintf("Hi %s, reset your password", name),
		Lines: []string{
			"We received a request to reset your SkillPass password. If this wasn't you, you can safely ignore this email.",
			"This link expires in 1 hour.",
		},
		CTAURL:   resetURL,
		CTALabel: "Reset password",
	})
}

// ApplicationReceivedEmail notifies a company that a candidate applied.
func ApplicationReceivedEmail(jobTitle, candidateName, link string) Message {
	return render(fmt.Sprintf("New application for %s", jobTitle), templateData{
		Heading: "You have a new application",
		Lines: []string{
			fmt.Sprintf("%s applied to your %q posting.", candidateName, jobTitle),
		},
		CTAURL:   link,
		CTALabel: "Review application",
	})
}

// StatusUpdateEmail notifies a jobseeker that their application moved.
func StatusUpdateEmail(jobTitle, status, link string) Message {
	return render(fmt.Sprintf("Update on your application for %s", jobTitle), templateData{
		Heading: "Your application status changed",
		Lines: []string{
			fmt.Sprintf("Your application for %q is now %q.", jobTitle, status),
		},
		CTAURL:   link,
		CTALabel: "View my applications",
	})
}

// NoteEmail notifies a jobseeker that the company left a note.
func NoteEmail(jobTitle, link string) Message {
	return render(fmt.Sprintf("New message about %s", jobTitle), templateData{
		Heading: "A company sent you a message",
		Lines: []string{
			fmt.Sprintf("The company left a note on your application for %q.", jobTitle),
		},
		CTAURL:   link,
		CTALabel: "Read the message",
	})
}
