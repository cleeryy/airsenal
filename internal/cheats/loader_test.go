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
	if err := s.Load(); err != nil {
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
	_ = s.Load()
	if len(s.List()) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(s.List()))
	}
}

func TestStore_Search(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "nmap.md", "---\ndescription: Port scanner\ntags: [network, recon]\n---\nnmap usage here")
	writeFile(t, dir, "curl.md", "---\ndescription: HTTP client\ntags: [http]\n---\ncurl usage here")

	s := NewStore(dir)
	_ = s.Load()

	results := s.Search("network")
	if len(results) != 1 || results[0].Topic != "nmap" {
		t.Fatalf("search 'network': got %v", topics(results))
	}

	results = s.Search("usage")
	if len(results) != 2 {
		t.Fatalf("search 'usage': expected 2, got %d", len(results))
	}
}

func TestStore_MissingDir(t *testing.T) {
	s := NewStore("/nonexistent/path")
	if err := s.Load(); err == nil {
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
