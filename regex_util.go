package sqlformatter

import (
	"regexp"
	"strings"
)

func EscapeRegExp(text string) string {
	return regexp.QuoteMeta(text)
}

func PatternToRegex(pattern string, caseInsensitive bool) *regexp.Regexp {
	prefix := "^"
	if caseInsensitive {
		prefix = "(?i)^"
	}
	return regexp.MustCompile(prefix + "(?:" + pattern + ")")
}

func ToCaseInsensitivePattern(prefix string) string {
	var b strings.Builder
	for _, r := range prefix {
		if r == ' ' {
			b.WriteString(`\s+`)
			continue
		}
		upper := strings.ToUpper(string(r))
		lower := strings.ToLower(string(r))
		if upper == lower {
			b.WriteString(regexp.QuoteMeta(string(r)))
		} else {
			b.WriteString("[")
			b.WriteString(upper)
			b.WriteString(lower)
			b.WriteString("]")
		}
	}
	return b.String()
}

func WithDashes(pattern string) string {
	return pattern + "(?:-" + pattern + ")*"
}

func PrefixesPattern(prefixes []string, requirePrefix bool) string {
	parts := make([]string, 0, len(prefixes)+1)
	for _, prefix := range prefixes {
		parts = append(parts, ToCaseInsensitivePattern(prefix))
	}
	if !requirePrefix {
		parts = append(parts, "")
	}
	return "(?:" + strings.Join(parts, "|") + ")"
}
