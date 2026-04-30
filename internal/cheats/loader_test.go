package cheats

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStore_LoadAndGet(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "nmap.md", "---\ndescription: Port scanner\ntags: [network]\n---\n\n## Usage\nnmap <target>\n")
	writeFile(t, dir, "curl.txt", "# curl\nHTTP client\n")
	writeFile(t, dir, "ignored.go", "package main") // must be ignored

	s := NewStore(dir)
	if _, err := s.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}

	cs, ok := s.Get("nmap")
	if !ok {
		t.Fatal("expected nmap to be loaded")
	}
	if cs.Description != "Port scanner" {
		t.Fatalf("nmap description: got %q", cs.Description)
	}
	if len(cs.Tags) != 1 || cs.Tags[0] != "network" {
		t.Fatalf("nmap tags: got %v", cs.Tags)
	}

	if _, ok := s.Get("curl"); !ok {
		t.Fatal("expected curl to be loaded")
	}
	if _, ok := s.Get("ignored"); ok {
		t.Fatal("ignored.go should not be loaded")
	}
}

func TestStore_List(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.md", "A")
	writeFile(t, dir, "b.md", "B")

	s := NewStore(dir)
	_, _ = s.Load()
	if len(s.List()) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(s.List()))
	}
}

func TestStore_SearchRanked_ranking(t *testing.T) {
	dir := t.TempDir()
	// "scan" appears in: nmap topic, nmap tags, curl description, ssh content
	writeFile(t, dir, "nmap.md", "---\ndescription: Port scanner\ntags: [scan, network]\n---\nnmap usage")
	writeFile(t, dir, "curl.md", "---\ndescription: scan-capable HTTP client\ntags: [http]\n---\ncurl usage")
	writeFile(t, dir, "ssh.md", "---\ndescription: Remote login\ntags: [remote]\n---\nssh scan example")

	s := NewStore(dir)
	_, _ = s.Load()

	results := s.SearchRanked("scan", 10)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// nmap matches by topic ("nmap" doesn't contain "scan") — wait, nmap has "scan" in tags
	// curl has "scan" in description; ssh has "scan" in content
	// Priority: tag(nmap) > description(curl) > content(ssh)
	if results[0].Topic != "nmap" || results[0].MatchReason != "tag" {
		t.Fatalf("rank 1: want nmap/tag, got %s/%s", results[0].Topic, results[0].MatchReason)
	}
	if results[1].Topic != "curl" || results[1].MatchReason != "description" {
		t.Fatalf("rank 2: want curl/description, got %s/%s", results[1].Topic, results[1].MatchReason)
	}
	if results[2].Topic != "ssh" || results[2].MatchReason != "content" {
		t.Fatalf("rank 3: want ssh/content, got %s/%s", results[2].Topic, results[2].MatchReason)
	}
}

func TestStore_SearchRanked_topicPriority(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "nmap.md", "---\ndescription: Port scanner\ntags: [network]\n---\nnmap usage")
	writeFile(t, dir, "curl.md", "---\ndescription: nmap-like tool\ntags: [http]\n---\ncurl usage")

	s := NewStore(dir)
	_, _ = s.Load()

	results := s.SearchRanked("nmap", 10)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Topic != "nmap" || results[0].MatchReason != "topic" {
		t.Fatalf("topic match should rank first, got %s/%s", results[0].Topic, results[0].MatchReason)
	}
}

func TestStore_SearchRanked_limit(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.md", "---\ntags: [test]\n---\nfoo")
	writeFile(t, dir, "b.md", "---\ntags: [test]\n---\nfoo")
	writeFile(t, dir, "c.md", "---\ntags: [test]\n---\nfoo")

	s := NewStore(dir)
	_, _ = s.Load()

	results := s.SearchRanked("test", 2)
	if len(results) != 2 {
		t.Fatalf("expected limit of 2 results, got %d", len(results))
	}
}

func TestStore_SearchRanked_emptySlice(t *testing.T) {
	s := NewStore(t.TempDir())
	_, _ = s.Load()
	results := s.SearchRanked("anything", 20)
	if results == nil {
		t.Fatal("expected non-nil empty slice, got nil")
	}
}

func TestStore_MissingDir(t *testing.T) {
	s := NewStore("/nonexistent/path")
	if _, err := s.Load(); err == nil {
		t.Fatal("expected error for missing directory")
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func topics(list []*Cheatsheet) []string {
	out := make([]string, len(list))
	for i, cs := range list {
		out[i] = cs.Topic
	}
	return out
}
