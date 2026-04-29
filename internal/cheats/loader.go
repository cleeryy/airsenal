package cheats

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Store is a thread-safe in-memory cache of loaded cheatsheets.
type Store struct {
	dir  string
	mu   sync.RWMutex
	data map[string]*Cheatsheet
}

// NewStore creates a Store backed by the given directory.
func NewStore(dir string) *Store {
	return &Store{dir: dir, data: make(map[string]*Cheatsheet)}
}

// Load reads all .md and .txt files from the configured directory into memory.
func (s *Store) Load() error {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return fmt.Errorf("reading cheats dir %q: %w", s.dir, err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]*Cheatsheet)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := filepath.Ext(e.Name())
		if ext != ".md" && ext != ".txt" {
			continue
		}
		topic := strings.TrimSuffix(e.Name(), ext)
		raw, err := os.ReadFile(filepath.Join(s.dir, e.Name()))
		if err != nil {
			continue
		}
		s.data[topic] = Parse(topic, string(raw))
	}
	return nil
}

// Get returns the cheatsheet for a topic (exact match, case-sensitive key).
func (s *Store) Get(topic string) (*Cheatsheet, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cs, ok := s.data[topic]
	return cs, ok
}

// List returns all loaded cheatsheets in no guaranteed order.
func (s *Store) List() []*Cheatsheet {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*Cheatsheet, 0, len(s.data))
	for _, cs := range s.data {
		list = append(list, cs)
	}
	return list
}

// SearchResult pairs a Cheatsheet with the reason it matched a query.
// MatchReason is one of: "topic", "tag", "description", "content".
type SearchResult struct {
	*Cheatsheet
	MatchReason string `json:"match_reason"`
}

// SearchRanked returns up to limit cheatsheets matching query, ranked by
// specificity: topic name match > tag match > description match > content match.
// Within each rank, results are sorted alphabetically by topic.
func (s *Store) SearchRanked(query string, limit int) []SearchResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	q := strings.ToLower(query)

	type candidate struct {
		cs     *Cheatsheet
		score  int
		reason string
	}
	var candidates []candidate
	for _, cs := range s.data {
		switch {
		case strings.Contains(strings.ToLower(cs.Topic), q):
			candidates = append(candidates, candidate{cs, 1, "topic"})
		case tagsContain(cs.Tags, q):
			candidates = append(candidates, candidate{cs, 2, "tag"})
		case strings.Contains(strings.ToLower(cs.Description), q):
			candidates = append(candidates, candidate{cs, 3, "description"})
		case strings.Contains(strings.ToLower(contentPreview(cs.Content, 512)), q):
			candidates = append(candidates, candidate{cs, 4, "content"})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score < candidates[j].score
		}
		return candidates[i].cs.Topic < candidates[j].cs.Topic
	})

	size := len(candidates)
	if size > limit {
		size = limit
	}
	out := make([]SearchResult, 0, size)
	for i, c := range candidates {
		if i >= limit {
			break
		}
		out = append(out, SearchResult{Cheatsheet: c.cs, MatchReason: c.reason})
	}
	return out
}

func tagsContain(tags []string, q string) bool {
	for _, t := range tags {
		if strings.Contains(strings.ToLower(t), q) {
			return true
		}
	}
	return false
}

func contentPreview(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
