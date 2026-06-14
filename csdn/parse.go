package csdn

import (
	"errors"
	"regexp"
	"strings"
)

var (
	reArticleURL = regexp.MustCompile(`csdn\.net/([^/]+)/article/details/(\d+)`)
	reShorthand  = regexp.MustCompile(`^([A-Za-z0-9_\-]+)/(\d+)$`)
	reDigits     = regexp.MustCompile(`^\d+$`)
	reUserURL    = regexp.MustCompile(`blog\.csdn\.net/([A-Za-z0-9_\-]+)`)
	reToken      = regexp.MustCompile(`^[A-Za-z0-9_\-]+$`)
	reReadCount  = regexp.MustCompile(`(\d+)`)

	reScriptStyle = regexp.MustCompile(`(?is)<(script|style)[^>]*>.*?</(script|style)>`)
	reTag         = regexp.MustCompile(`(?s)<[^>]+>`)
	reBlankLines  = regexp.MustCompile(`\n{3,}`)
	reEm          = regexp.MustCompile(`(?i)</?em>`)
)

// reserved are the path segments under blog.csdn.net that are not usernames.
var reserved = map[string]bool{
	"article":   true,
	"phoenix":   true,
	"community": true,
	"download":  true,
	"nav":       true,
	"so":        true,
	"rank":      true,
}

// ParseArticleRef reads the author username and numeric id of an article from a
// full url (https://blog.csdn.net/{user}/article/details/{id}) or the
// shorthand username/123456. A bare numeric id is rejected: CSDN article urls
// carry the author username, so an id alone cannot be located.
func ParseArticleRef(in string) (username, id string, err error) {
	in = strings.TrimSpace(in)
	if in == "" {
		return "", "", errors.New("empty article reference")
	}
	if m := reArticleURL.FindStringSubmatch(in); m != nil {
		return m[1], m[2], nil
	}
	if m := reShorthand.FindStringSubmatch(in); m != nil {
		return m[1], m[2], nil
	}
	if reDigits.MatchString(in) {
		return "", "", errors.New("a bare article id is ambiguous on CSDN — pass the full url or username/id")
	}
	return "", "", errors.New("could not read an article reference from " + in)
}

// ParseArticleID reads just the numeric article id from a full url, the
// username/id shorthand, or a bare id. Unlike ParseArticleRef it accepts a bare
// id, because the surfaces keyed only by id (such as the comment list) do not
// need the author username.
func ParseArticleID(in string) (string, error) {
	in = strings.TrimSpace(in)
	if in == "" {
		return "", errors.New("empty article reference")
	}
	if m := reArticleURL.FindStringSubmatch(in); m != nil {
		return m[2], nil
	}
	if m := reShorthand.FindStringSubmatch(in); m != nil {
		return m[2], nil
	}
	if reDigits.MatchString(in) {
		return in, nil
	}
	return "", errors.New("could not read an article id from " + in)
}

// ParseUsername reads a CSDN username from a profile url
// (https://blog.csdn.net/{username}) or a bare token. Reserved path segments
// such as article or phoenix are rejected.
func ParseUsername(in string) (string, error) {
	in = strings.TrimSpace(in)
	if in == "" {
		return "", errors.New("empty username")
	}
	if m := reUserURL.FindStringSubmatch(in); m != nil {
		name := m[1]
		if reserved[name] {
			return "", errors.New("not a username: " + name)
		}
		return name, nil
	}
	if strings.ContainsAny(in, "/ ?#:") {
		return "", errors.New("could not read a username from " + in)
	}
	if !reToken.MatchString(in) {
		return "", errors.New("could not read a username from " + in)
	}
	if reserved[in] {
		return "", errors.New("not a username: " + in)
	}
	return in, nil
}

// stripEm removes the <em>/</em> highlight tags CSDN search wraps around matched
// terms in titles and descriptions.
func stripEm(s string) string {
	return reEm.ReplaceAllString(s, "")
}

// stripTags turns an HTML fragment into readable plain text: it drops
// <script>/<style> blocks whole, strips remaining tags, unescapes the common
// entities, and collapses runs of blank lines.
func stripTags(html string) string {
	s := reScriptStyle.ReplaceAllString(html, "")
	s = reTag.ReplaceAllString(s, "")
	s = unescapeHTML(s)
	s = reBlankLines.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}

// unescapeHTML decodes the handful of HTML entities the CSDN body carries.
func unescapeHTML(s string) string {
	r := strings.NewReplacer(
		"&amp;", "&",
		"&lt;", "<",
		"&gt;", ">",
		"&quot;", `"`,
		"&#39;", "'",
		"&#x27;", "'",
		"&nbsp;", " ",
	)
	return r.Replace(s)
}

// extractBalancedJSON returns the JSON object that begins at the first { at or
// after start, scanning for its matching close brace while respecting strings
// and escapes. It reports false when no balanced object is found.
func extractBalancedJSON(s string, start int) (string, bool) {
	i := start
	for i < len(s) && s[i] != '{' {
		i++
	}
	if i >= len(s) {
		return "", false
	}
	depth := 0
	inStr := false
	esc := false
	for j := i; j < len(s); j++ {
		c := s[j]
		if inStr {
			switch {
			case esc:
				esc = false
			case c == '\\':
				esc = true
			case c == '"':
				inStr = false
			}
			continue
		}
		switch c {
		case '"':
			inStr = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[i : j+1], true
			}
		}
	}
	return "", false
}

// firstInt returns the first run of digits in s as a string, or "".
func firstInt(s string) string {
	if m := reReadCount.FindString(s); m != "" {
		return m
	}
	return ""
}
