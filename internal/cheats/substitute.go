package cheats

import (
	"regexp"
	"sort"
	"strings"
)

var varPattern = regexp.MustCompile(`<([a-zA-Z][a-zA-Z0-9_-]*)>`)

// ExtractVariables returns sorted, deduplicated variable names found in content.
func ExtractVariables(content string) []string {
	seen := map[string]bool{}
	for _, m := range varPattern.FindAllStringSubmatch(content, -1) {
		seen[m[1]] = true
	}
	vars := make([]string, 0, len(seen))
	for v := range seen {
		vars = append(vars, v)
	}
	sort.Strings(vars)
	return vars
}

// Substitute replaces <key> placeholders with the corresponding values.
// When decorate is true, each substituted value is wrapped in ANSI bold.
func Substitute(content string, vars map[string]string, decorate bool) string {
	for k, v := range vars {
		var replacement string
		if decorate {
			replacement = "\x1b[1m" + v + "\x1b[0m"
		} else {
			replacement = v
		}
		content = strings.ReplaceAll(content, "<"+k+">", replacement)
	}
	return content
}
