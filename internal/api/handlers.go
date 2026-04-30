package api

import (
	"encoding/json"
	"fmt"
	"log"
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
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		log.Printf("json encode error: %v", err)
	}
}

// list handles GET /
func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	all := h.store.List()
	sort.Slice(all, func(i, j int) bool { return all[i].Topic < all[j].Topic })

	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(all); err != nil {
			log.Printf("json encode error: %v", err)
		}
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

func (h *handler) random(w http.ResponseWriter, r *http.Request) {
	cat := r.URL.Query().Get("cat")
	tag := r.URL.Query().Get("tag")

	var cs *cheats.Cheatsheet
	if cat != "" {
		cs = h.store.RandomByCategory(cat)
		if cs == nil {
			http.Error(w, fmt.Sprintf("no random cheatsheet found in category %q\n", cat), http.StatusNotFound)
			return
		}
	} else if tag != "" {
		cs = h.store.RandomByTag(tag)
		if cs == nil {
			http.Error(w, fmt.Sprintf("no random cheatsheet found with tag %q\n", tag), http.StatusNotFound)
			return
		}
	} else {
		cs = h.store.Random()
		if cs == nil {
			http.Error(w, "no cheatsheets found in store\n", http.StatusNotFound)
			return
		}
	}

	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		type randomResponse struct {
			*cheats.Cheatsheet
			Random bool `json:"random"`
		}
		json.NewEncoder(w).Encode(randomResponse{Cheatsheet: cs, Random: true})
		return
	}

	vars := templateVars(r)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	terminal := isTerminalClient(r)
	header := fmt.Sprintf("🎲 Random cheat: %s [%s]", cs.Topic, cs.Category)
	if cs.Category == "" {
		header = fmt.Sprintf("🎲 Random cheat: %s", cs.Topic)
	}

	if terminal {
		sep := strings.Repeat("━", len([]rune(header)))
		fmt.Fprintf(w, "\x1b[1;36m%s\n%s\x1b[0m\n", header, sep)
	} else {
		sep := strings.Repeat("=", len(header))
		fmt.Fprintf(w, "%s\n%s\n\n", header, sep)
	}

	fmt.Fprint(w, render.Render(cs, vars, terminal))
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
		if err := json.NewEncoder(w).Encode(results); err != nil {
			log.Printf("json encode error: %v", err)
		}
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

func (h *handler) listCategories(w http.ResponseWriter, r *http.Request) {
	cats := h.store.ListCategories()

	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(cats); err != nil {
			log.Printf("json encode error: %v", err)
		}
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	title := fmt.Sprintf("Available categories (%d)", len(cats))
	fmt.Fprintf(w, "%s\n", title)
	fmt.Fprintf(w, "%s\n", strings.Repeat("─", len([]rune(title))))
	for _, c := range cats {
		fmt.Fprintf(w, "%-16s (%2d cheats)\n", c.Category, c.Count)
	}
	fmt.Fprintf(w, "\nUsage: curl cheat.example.com/~cat/<category>\n")
}

func (h *handler) listByCategory(w http.ResponseWriter, r *http.Request) {
	cat := r.PathValue("category")
	cheats := h.store.ListByCategory(cat)

	if len(cheats) == 0 {
		http.Error(w, fmt.Sprintf("Category not found: %q\n", cat), http.StatusNotFound)
		return
	}

	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(cheats); err != nil {
			log.Printf("json encode error: %v", err)
		}
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	displayCat := cheats[0].Category
	if displayCat == "" {
		displayCat = cat
	}
	title := fmt.Sprintf("Category: %s (%d cheats)", displayCat, len(cheats))
	fmt.Fprintf(w, "%s\n", title)
	fmt.Fprintf(w, "%s\n", strings.Repeat("─", len([]rune(title))))
	for _, cs := range cheats {
		if cs.Description != "" {
			fmt.Fprintf(w, "%-16s %s\n", cs.Topic, cs.Description)
		} else {
			fmt.Fprintf(w, "%s\n", cs.Topic)
		}
	}
	fmt.Fprintf(w, "\nUsage: curl cheat.example.com/<topic>\n")
}

func (h *handler) listTags(w http.ResponseWriter, r *http.Request) {
	tags := h.store.ListTags()

	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(tags); err != nil {
			log.Printf("json encode error: %v", err)
		}
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	title := fmt.Sprintf("Available tags (%d)", len(tags))
	fmt.Fprintf(w, "%s\n", title)
	fmt.Fprintf(w, "%s\n", strings.Repeat("─", len([]rune(title))))
	for _, t := range tags {
		fmt.Fprintf(w, "%-16s (%2d)\n", t.Tag, t.Count)
	}
	fmt.Fprintf(w, "\nUsage: curl cheat.example.com/~tag/<tag>\n")
}

func (h *handler) listByTag(w http.ResponseWriter, r *http.Request) {
	tag := r.PathValue("tag")
	cheats := h.store.ListByTag(tag)

	if len(cheats) == 0 {
		http.Error(w, fmt.Sprintf("Tag not found: %q\n", tag), http.StatusNotFound)
		return
	}

	if wantsJSON(r) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(cheats); err != nil {
			log.Printf("json encode error: %v", err)
		}
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	title := fmt.Sprintf("Tag: %s (%d cheats)", tag, len(cheats))
	fmt.Fprintf(w, "%s\n", title)
	fmt.Fprintf(w, "%s\n", strings.Repeat("─", len([]rune(title))))
	for _, cs := range cheats {
		if cs.Description != "" {
			fmt.Fprintf(w, "%-16s %s\n", cs.Topic, cs.Description)
		} else {
			fmt.Fprintf(w, "%s\n", cs.Topic)
		}
	}
	fmt.Fprintf(w, "\nUsage: curl cheat.example.com/<topic>\n")
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
