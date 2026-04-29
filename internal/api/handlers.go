package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/cleeryy/airsenal/internal/cheats"
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

	if r.URL.Query().Get("raw") == "1" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprint(w, cs.Raw)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, cs.Content)
}

// wantsJSON reports whether the client prefers a JSON response.
func wantsJSON(r *http.Request) bool {
	return r.URL.Query().Get("format") == "json" ||
		strings.Contains(r.Header.Get("Accept"), "application/json")
}
