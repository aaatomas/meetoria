package organization

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var consecutiveHyphens = regexp.MustCompile(`-+`)

func Slugify(value string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	normalized, _, _ := transform.String(t, strings.TrimSpace(value))

	var b strings.Builder
	lastHyphen := false
	for _, r := range strings.ToLower(normalized) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastHyphen = false
			continue
		}
		if !lastHyphen && b.Len() > 0 {
			b.WriteByte('-')
			lastHyphen = true
		}
	}

	slug := consecutiveHyphens.ReplaceAllString(strings.Trim(b.String(), "-"), "-")
	return slug
}
