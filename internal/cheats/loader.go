package cheats

import (
	"fmt"
	"os"
	"path/filepath"
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
