package render

import (
	"regexp"
	"strings"

	"github.com/cleeryy/airsenal/internal/cheats"
)

const (
	reset   = "\x1b[0m"
	bold    = "\x1b[1m"
	dim     = "\x1b[2m"
	cyan    = "\x1b[36m"
	yellow  = "\x1b[33m"
	white   = "\x1b[97m"
	green   = "\x1b[32m"
	magenta = "\x1b[35m"
	gray    = "\x1b[90m"
)

var varRe = regexp.MustCompile(`<([a-zA-Z][a-zA-Z0-9_-]*)>`)

// Render formats cs for output. When terminal is true ANSI escape codes are applied;
// otherwise plain text with bold substitution decoration is returned.
func Render(cs *cheats.Cheatsheet, vars map[string]string, terminal bool) string {
	if !terminal {
		return renderPlain(cs, vars)
	}
	return renderANSI(cs, vars)
}

func renderPlain(cs *cheats.Cheatsheet, vars map[string]string) string {
	content := cs.Content
	if len(vars) > 0 {
		content = cheats.Substitute(content, vars, true)
	}
	if available := cheats.ExtractVariables(cs.Content); len(available) > 0 {
		display := make([]string, len(available))
		for i, v := range available {
			display[i] = "<" + v + ">"
		}
		content += "\nVariables: " + strings.Join(display, "  ") + "\n"
	}
	return content
}

func renderANSI(cs *cheats.Cheatsheet, vars map[string]string) string {
	var sb strings.Builder
	inCodeBlock := false

	for _, line := range strings.Split(cs.Content, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		if inCodeBlock {
			sb.WriteString(bold + white + colorVars(line, vars, bold+white) + reset + "\n")
			continue
		}

		if trimmed == "---" || trimmed == "***" || trimmed == "___" {
			continue
		}

		if strings.HasPrefix(trimmed, "## ") {
			sb.WriteString(bold + yellow + strings.TrimPrefix(trimmed, "## ") + reset + "\n")
			continue
		}

		if strings.HasPrefix(trimmed, "# ") {
			heading := strings.TrimPrefix(trimmed, "# ")
			sep := strings.Repeat("━", len([]rune(heading)))
			sb.WriteString(bold + cyan + colorVars(heading, vars, bold+cyan) + reset + "\n")
			sb.WriteString(cyan + sep + reset + "\n")
			continue
		}

		if strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "\t") {
			sb.WriteString(renderCmd(line, vars) + "\n")
			continue
		}

		if trimmed == "" {
			sb.WriteByte('\n')
			continue
		}

		sb.WriteString(colorVars(line, vars, "") + "\n")
	}

	if available := cheats.ExtractVariables(cs.Content); len(available) > 0 {
		sb.WriteString("\n" + dim + "Variables:" + reset)
		for _, v := range available {
			if val, ok := vars[v]; ok {
				sb.WriteString("  " + bold + magenta + val + reset)
			} else {
				sb.WriteString("  " + bold + green + "<" + v + ">" + reset)
			}
		}
		sb.WriteByte('\n')
	}

	return sb.String()
}

func renderCmd(line string, vars map[string]string) string {
	idx := strings.Index(line, " # ")
	if idx >= 0 {
		return bold + white + colorVars(line[:idx], vars, bold+white) + reset +
			gray + line[idx:] + reset
	}
	return bold + white + colorVars(line, vars, bold+white) + reset
}

// colorVars replaces variable placeholders with ANSI-colored versions.
// restore is the escape sequence to re-apply after each replacement (e.g. bold+white inside a command).
func colorVars(text string, vars map[string]string, restore string) string {
	return varRe.ReplaceAllStringFunc(text, func(match string) string {
		name := match[1 : len(match)-1]
		if val, ok := vars[name]; ok {
			return reset + bold + magenta + val + reset + restore
		}
		return reset + bold + green + match + reset + restore
	})
}
