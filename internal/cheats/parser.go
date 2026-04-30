package cheats

import (
	"bufio"
	"strings"
)

// Parse creates a Cheatsheet from raw file content.
// Files may optionally begin with a YAML-style frontmatter block delimited by "---".
func Parse(topic, raw string) *Cheatsheet {
	cs := &Cheatsheet{Topic: topic, Raw: raw}

	if strings.HasPrefix(raw, "---\n") {
		body, meta := parseFrontmatter(raw)
		cs.Description = meta["description"]
		cs.Category = meta["category"]
		if tagStr, ok := meta["tags"]; ok {
			tagStr = strings.Trim(tagStr, "[]")
			for _, t := range strings.Split(tagStr, ",") {
				if t = strings.TrimSpace(t); t != "" {
					cs.Tags = append(cs.Tags, t)
				}
			}
		}
		cs.Content = body
		if cs.Tags == nil {
			cs.Tags = []string{}
		}
		return cs
	}

	cs.Content = raw
	cs.Description = extractDescription(raw)
	if cs.Tags == nil {
		cs.Tags = []string{}
	}
	return cs
}

// parseFrontmatter splits "---\n<meta>\n---\n<body>" and returns the body plus parsed key-value pairs.
func parseFrontmatter(raw string) (string, map[string]string) {
	meta := make(map[string]string)
	without := raw[4:] // strip leading "---\n"
	end := strings.Index(without, "\n---\n")
	if end == -1 {
		return raw, meta
	}
	block := without[:end]
	body := strings.TrimSpace(without[end+5:])
	scanner := bufio.NewScanner(strings.NewReader(block))
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), ":", 2)
		if len(parts) == 2 {
			meta[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return body, meta
}

// extractDescription pulls the first meaningful text line from content.
func extractDescription(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
		if !strings.HasPrefix(line, "#") {
			return line
		}
	}
	return ""
}
