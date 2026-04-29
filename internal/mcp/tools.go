package mcp

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/cleeryy/airsenal/internal/cheats"
)

// tools wires MCP tool definitions to the cheatsheet store.
type tools struct {
	store *cheats.Store
}

func newTools(store *cheats.Store) *tools {
	return &tools{store: store}
}

// definitions returns the static list of available MCP tools.
func (t *tools) definitions() []toolDef {
	return []toolDef{
		{
			Name:        "list_cheatsheets",
			Description: "List all available cheatsheet topics with their descriptions.",
			InputSchema: inputSchema{Type: "object"},
		},
		{
			Name:        "get_cheatsheet",
			Description: "Retrieve the full content and metadata for a specific cheatsheet topic.",
			InputSchema: inputSchema{
				Type: "object",
				Properties: map[string]property{
					"topic": {Type: "string", Description: "The cheatsheet topic name (e.g. nmap, curl, ssh)"},
				},
				Required: []string{"topic"},
			},
		},
		{
			Name:        "search_cheatsheets",
			Description: "Search cheatsheets by keyword across topic names, descriptions, tags, and content.",
			InputSchema: inputSchema{
				Type: "object",
				Properties: map[string]property{
					"query": {Type: "string", Description: "Search keyword or phrase"},
				},
				Required: []string{"query"},
			},
		},
	}
}

// call dispatches a tool invocation and returns the result.
func (t *tools) call(name string, rawArgs json.RawMessage) toolResult {
	switch name {
	case "list_cheatsheets":
		return t.listCheatsheets()
	case "get_cheatsheet":
		return t.getCheatsheet(rawArgs)
	case "search_cheatsheets":
		return t.searchCheatsheets(rawArgs)
	default:
		return errResult(fmt.Sprintf("unknown tool: %s", name))
	}
}

func (t *tools) listCheatsheets() toolResult {
	all := t.store.List()
	sort.Slice(all, func(i, j int) bool { return all[i].Topic < all[j].Topic })

	var sb strings.Builder
	fmt.Fprintf(&sb, "Available cheatsheets (%d topics):\n\n", len(all))
	for _, cs := range all {
		if cs.Description != "" {
			fmt.Fprintf(&sb, "  %-16s %s\n", cs.Topic, cs.Description)
		} else {
			fmt.Fprintf(&sb, "  %s\n", cs.Topic)
		}
	}
	return textResult(sb.String())
}

func (t *tools) getCheatsheet(rawArgs json.RawMessage) toolResult {
	var args struct {
		Topic string `json:"topic"`
	}
	if err := json.Unmarshal(rawArgs, &args); err != nil {
		return errResult("invalid arguments: " + err.Error())
	}
	topic := strings.ToLower(strings.TrimSpace(args.Topic))
	if topic == "" {
		return errResult("topic is required")
	}
	cs, ok := t.store.Get(topic)
	if !ok {
		return errResult(fmt.Sprintf("no cheatsheet found for %q", topic))
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "Topic: %s\n", cs.Topic)
	if cs.Description != "" {
		fmt.Fprintf(&sb, "Description: %s\n", cs.Description)
	}
	if len(cs.Tags) > 0 {
		fmt.Fprintf(&sb, "Tags: %s\n", strings.Join(cs.Tags, ", "))
	}
	fmt.Fprintf(&sb, "\n%s\n", cs.Content)
	return textResult(sb.String())
}

func (t *tools) searchCheatsheets(rawArgs json.RawMessage) toolResult {
	var args struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal(rawArgs, &args); err != nil {
		return errResult("invalid arguments: " + err.Error())
	}
	q := strings.TrimSpace(args.Query)
	if q == "" {
		return errResult("query is required")
	}
	results := t.store.SearchRanked(q, 20)
	if len(results) == 0 {
		return textResult(fmt.Sprintf("No cheatsheets matched %q.", q))
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "Search results for %q (%d match(es)):\n\n", q, len(results))
	for _, r := range results {
		fmt.Fprintf(&sb, "  %-16s %s [%s]\n", r.Topic, r.Description, r.MatchReason)
		if r.MatchReason == "tag" && len(r.Tags) > 0 {
			fmt.Fprintf(&sb, "  %16s tags: %s\n", "", strings.Join(r.Tags, ", "))
		}
	}
	return textResult(sb.String())
}

func textResult(text string) toolResult {
	return toolResult{Content: []contentItem{{Type: "text", Text: text}}}
}

func errResult(msg string) toolResult {
	return toolResult{IsError: true, Content: []contentItem{{Type: "text", Text: msg}}}
}
