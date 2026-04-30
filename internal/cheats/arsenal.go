package cheats

import (
	"bufio"
	"strings"
)

// ParseArsenal parses a single Arsenal-format markdown file into a Cheatsheet.
//
// Arsenal markdown format:
//
//	# Tool Name            ← H1: file-level title → Topic
//	% tag1, tag2           ← file-level tags (% prefix, comma-separated)
//	#key/value ...         ← file-level command tags (applied to all commands)
//
//	## Command Name        ← H2+: individual command entry
//	#key/value ...         ← per-command tag line (optional)
//	description text       ← free-text description
//	```
//	command <variable>
//	```
//
// The topic is the filename stem (caller's responsibility to pass it).
// category is the relative directory path from the arsenal root (e.g. "Scan", "Active_directory/Impacket").
// raw is the original file content.
func ParseArsenal(topic, category, raw string) *Cheatsheet {
	cs := &Cheatsheet{
		Topic:    topic,
		Category: category,
		Tags:     []string{},
		Raw:      raw,
	}

	var (
		contentParts   []string
		inCode         bool
		codeLines      []string
		currentSection string
		descLines      []string
	)

	flushSection := func() {
		if currentSection == "" && len(descLines) == 0 && len(codeLines) == 0 {
			return
		}
		var sb strings.Builder
		if currentSection != "" {
			sb.WriteString("## ")
			sb.WriteString(currentSection)
			sb.WriteString("\n")
		}
		for _, d := range descLines {
			sb.WriteString(d)
			sb.WriteString("\n")
		}
		if len(codeLines) > 0 {
			sb.WriteString("```\n")
			for _, cl := range codeLines {
				sb.WriteString(cl)
				sb.WriteString("\n")
			}
			sb.WriteString("```\n")
		}
		contentParts = append(contentParts, sb.String())
		currentSection = ""
		descLines = nil
		codeLines = nil
	}

	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()

		if inCode {
			if strings.HasPrefix(line, "```") {
				inCode = false
			} else {
				codeLines = append(codeLines, line)
			}
			continue
		}

		if strings.HasPrefix(line, "```") {
			inCode = true
			continue
		}

		if strings.HasPrefix(line, "# ") && !strings.HasPrefix(line, "## ") {
			title := strings.TrimPrefix(line, "# ")
			if cs.Description == "" {
				cs.Description = title
			}
			continue
		}

		if strings.HasPrefix(line, "## ") {
			flushSection()
			currentSection = strings.TrimSpace(line[3:])
			continue
		}

		if strings.HasPrefix(line, "%") {
			tagStr := strings.TrimSpace(line[1:])
			for _, t := range strings.Split(tagStr, ",") {
				if t = strings.TrimSpace(t); t != "" {
					cs.Tags = append(cs.Tags, t)
				}
			}
			continue
		}

		if isCommandTagLine(line) {
			continue
		}

		if (strings.HasPrefix(line, "$ ") || strings.HasPrefix(line, "= ")) && strings.Contains(line, ":") {
			continue
		}

		if strings.TrimSpace(line) == "" {
			if currentSection != "" && (len(descLines) > 0 || len(codeLines) > 0) {
				descLines = append(descLines, "")
			}
			continue
		}

		descLines = append(descLines, line)
	}

	flushSection()

	cs.Content = strings.TrimSpace(strings.Join(contentParts, "\n"))
	return cs
}

func isCommandTagLine(line string) bool {
	if !strings.HasPrefix(line, "#") {
		return false
	}
	if len(line) < 2 {
		return false
	}
	c := line[1]
	return c != ' ' && c != '#' && c != '\t'
}
