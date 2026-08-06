package lib

import "net/url"

// RedactURL strips credentials and tokens from a meeting/calendar link by
// removing its query string and fragment. It returns the scheme://host/path
// prefix when the input parses as a URL, and for scheme-less inputs (e.g.
// "meet.google.com/xyz?token=…") it cuts everything from the first '?' or '#'.
// On parse failure the input is returned unchanged rather than mangled.
func RedactURL(raw string) string {
	if raw == "" {
		return raw
	}
	// Scheme-less URLs (e.g. "meet.google.com/abc?pwd=x") still carry query
	// credentials but don't parse as absolute URLs — strip at the first
	// query/fragment marker for those.
	if !containsURLScheme(raw) {
		return cutAtQuery(raw)
	}
	u, err := url.Parse(raw)
	if err != nil {
		return cutAtQuery(raw)
	}
	out := u.Scheme + "://" + u.Host + u.Path
	if out == "://" {
		return cutAtQuery(raw)
	}
	return out
}

func containsURLScheme(s string) bool {
	for i := 0; i+2 < len(s); i++ {
		if s[i] == ':' && s[i+1] == '/' && s[i+2] == '/' {
			return true
		}
	}
	return false
}

func cutAtQuery(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == '?' || s[i] == '#' {
			return s[:i]
		}
	}
	return s
}
