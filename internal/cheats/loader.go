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

// Search returns cheatsheets whose topic, description, tags, or content preview
// contain the given query (case-insensitive substring match).
func (s *Store) Search(query string) []*Cheatsheet {
	s.mu.RLock()
	defer s.mu.RUnlock()
	q := strings.ToLower(query)
	var results []*Cheatsheet
	for _, cs := range s.data {
		if strings.Contains(strings.ToLower(cs.Topic), q) ||
			strings.Contains(strings.ToLower(cs.Description), q) ||
			tagsContain(cs.Tags, q) ||
			strings.Contains(strings.ToLower(contentPreview(cs.Content, 512)), q) {
			results = append(results, cs)
		}
	}
	return results
}

// SearchRanked returns up to limit cheatsheets matching query, ranked by
// specificity: topic name match > tag match > description match > content match.
// Within each rank, results are sorted alphabetically by topic.
func (s *Store) SearchRanked(query string, limit int) []*Cheatsheet {
	s.mu.RLock()
	defer s.mu.RUnlock()
	q := strings.ToLower(query)

	type scored struct {
		cs    *Cheatsheet
		score int
	}
	var results []scored
	for _, cs := range s.data {
		switch {
		case strings.Contains(strings.ToLower(cs.Topic), q):
			results = append(results, scored{cs, 1})
		case tagsContain(cs.Tags, q):
			results = append(results, scored{cs, 2})
		case strings.Contains(strings.ToLower(cs.Description), q):
			results = append(results, scored{cs, 3})
		case strings.Contains(strings.ToLower(contentPreview(cs.Content, 512)), q):
			results = append(results, scored{cs, 4})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].score != results[j].score {
			return results[i].score < results[j].score
		}
		return results[i].cs.Topic < results[j].cs.Topic
	})

	size := len(results)
	if size > limit {
		size = limit
	}
	out := make([]*Cheatsheet, 0, size)
	for i, r := range results {
		if i >= limit {
			break
		}
		out = append(out, r.cs)
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
