package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/cleeryy/airsenal/internal/cheats"
	"github.com/cleeryy/airsenal/internal/render"
)

type handler struct {
	store *cheats.Store
}

func newHandler(store *cheats.Store) *handler {
	return &handler{store: store}
}

// health handles GET /healthz
func (h *handler) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// list handles GET /
func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	all := h.store.List()
	sort.Slice(all, func(i, j int) bool { return all[i].Topic < all[j].Topic })

	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(all)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "Available cheatsheets (%d topics):\n\n", len(all))
	for _, cs := range all {
		if cs.Description != "" {
			fmt.Fprintf(w, "  %-16s %s\n", cs.Topic, cs.Description)
		} else {
			fmt.Fprintf(w, "  %s\n", cs.Topic)
		}
	}
	fmt.Fprintf(w, "\nUsage: curl cheat.example.com/<topic>\n")
}

// getTopic handles GET /{topic}
func (h *handler) getTopic(w http.ResponseWriter, r *http.Request) {
	topic := strings.ToLower(r.PathValue("topic"))

	cs, ok := h.store.Get(topic)
	if !ok {
		http.Error(w, fmt.Sprintf("no cheatsheet found for %q\n", topic), http.StatusNotFound)
		return
	}

	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cs)
		return
	}

	vars := templateVars(r)

	if r.URL.Query().Get("raw") == "1" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		content := cs.Raw
		if len(vars) > 0 {
			content = cheats.Substitute(content, vars, false)
		}
		fmt.Fprint(w, content)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, render.Render(cs, vars, isTerminalClient(r)))
}

// search handles GET /~search?q=<query>[&limit=N][&format=json]
func (h *handler) search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "missing required parameter: q\n", http.StatusBadRequest)
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	results := h.store.SearchRanked(q, limit)

	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if len(results) == 0 {
		fmt.Fprintf(w, "No results for %q\n", q)
		return
	}
	noun := "matches"
	if len(results) == 1 {
		noun = "match"
	}
	fmt.Fprintf(w, "Search results for %q (%d %s):\n\n", q, len(results), noun)
	for _, sr := range results {
		ann := searchAnnotation(sr)
		if ann != "" {
			fmt.Fprintf(w, "%s %s %s\n", sr.Topic, sr.Description, ann)
		} else {
			fmt.Fprintf(w, "%s %s\n", sr.Topic, sr.Description)
		}
	}
	fmt.Fprintf(w, "\nUsage: curl cheat.example.com/<topic>\n")
}

// searchAnnotation returns the bracketed suffix shown after a result line in plain-text output.
// Topic matches are self-evident; other match reasons are annotated for context.
func searchAnnotation(sr cheats.SearchResult) string {
	switch sr.MatchReason {
	case "tag":
		return "[tags: " + strings.Join(sr.Tags, ", ") + "]"
	case "description":
		return "[description match]"
	case "content":
		return "[content match]"
	default:
		return ""
	}
}

// isTerminalClient reports whether the request comes from a terminal-oriented HTTP client.
func isTerminalClient(r *http.Request) bool {
	ua := strings.ToLower(r.Header.Get("User-Agent"))
	return strings.Contains(ua, "curl") ||
		strings.Contains(ua, "wget") ||
		strings.Contains(ua, "httpie") ||
		strings.Contains(ua, "go-http-client")
}

// templateVars extracts non-reserved query parameters for placeholder substitution.
func templateVars(r *http.Request) map[string]string {
	reserved := map[string]bool{"raw": true, "format": true}
	vars := map[string]string{}
	for k, vals := range r.URL.Query() {
		if reserved[k] || len(vals) == 0 {
			continue
		}
		vars[k] = vals[0]
	}
	return vars
}

// wantsJSON reports whether the client prefers a JSON response.
func wantsJSON(r *http.Request) bool {
	return r.URL.Query().Get("format") == "json" ||
		strings.Contains(r.Header.Get("Accept"), "application/json")
}
