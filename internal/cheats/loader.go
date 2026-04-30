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

func (s *Store) Load() (int, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return 0, fmt.Errorf("reading cheats dir %q: %w", s.dir, err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]*Cheatsheet)
	count := 0
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
		count++
	}
	return count, nil
}

func (s *Store) LoadArsenal(arsenalDir string) (int, error) {
	if arsenalDir == "" {
		return 0, nil
	}

	info, err := os.Stat(arsenalDir)
	if err != nil || !info.IsDir() {
		return 0, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	err = filepath.WalkDir(arsenalDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && d.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Ext(path) != ".md" {
			return nil
		}

		rel, err := filepath.Rel(arsenalDir, path)
		if err != nil {
			return nil
		}
		category := filepath.Dir(rel)
		if category == "." {
			category = ""
		}

		topic := strings.TrimSuffix(filepath.Base(path), ".md")
		topic = strings.ToLower(topic)

		raw, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		cs := ParseArsenal(topic, category, string(raw))

		if _, exists := s.data[topic]; !exists {
			s.data[topic] = cs
			count++
		}

		return nil
	})

	return count, err
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

// ListCategories returns all available categories with their cheat count, sorted alphabetically.
func (s *Store) ListCategories() []CategorySummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	counts := make(map[string]int)
	displays := make(map[string]string)
	for _, cs := range s.data {
		if cs.Category != "" {
			catLower := strings.ToLower(cs.Category)
			counts[catLower]++
			if displays[catLower] == "" {
				displays[catLower] = cs.Category
			}
		}
	}

	var list []CategorySummary
	for catLower, count := range counts {
		list = append(list, CategorySummary{Category: displays[catLower], Count: count})
	}

	sort.Slice(list, func(i, j int) bool {
		return strings.ToLower(list[i].Category) < strings.ToLower(list[j].Category)
	})

	return list
}

// ListByCategory returns all cheats in a given category (case-insensitive), sorted alphabetically.
func (s *Store) ListByCategory(category string) []*Cheatsheet {
	s.mu.RLock()
	defer s.mu.RUnlock()

	catLower := strings.ToLower(category)
	var list []*Cheatsheet
	for _, cs := range s.data {
		if strings.ToLower(cs.Category) == catLower {
			list = append(list, cs)
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Topic < list[j].Topic
	})

	return list
}

func (s *Store) ListTags() []TagSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	counts := make(map[string]int)
	for _, cs := range s.data {
		for _, tag := range cs.Tags {
			counts[strings.ToLower(tag)]++
		}
	}

	var list []TagSummary
	for tag, count := range counts {
		list = append(list, TagSummary{Tag: tag, Count: count})
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].Count != list[j].Count {
			return list[i].Count > list[j].Count
		}
		return list[i].Tag < list[j].Tag
	})

	return list
}

// ListByTag returns all cheats that have the given tag (case-insensitive), sorted alphabetically.
func (s *Store) ListByTag(tag string) []*Cheatsheet {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tagLower := strings.ToLower(tag)
	var list []*Cheatsheet
	for _, cs := range s.data {
		for _, t := range cs.Tags {
			if strings.ToLower(t) == tagLower {
				list = append(list, cs)
				break
			}
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Topic < list[j].Topic
	})

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
