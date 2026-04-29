package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cleeryy/airsenal/internal/cheats"
)

func newTestStore(t *testing.T) *cheats.Store {
	t.Helper()
	dir := t.TempDir()
	content := "---\ndescription: Port scanner\ntags: [network]\n---\n\n## Usage\nnmap <target>\n"
	if err := os.WriteFile(filepath.Join(dir, "nmap.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	s := cheats.NewStore(dir)
	if err := s.Load(); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestHealthz(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/healthz", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Fatalf("body: %v", body)
	}
}

func TestList_plainText(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "nmap") {
		t.Fatalf("expected nmap in list, got: %s", body)
	}
}

func TestList_JSON(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/?format=json", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d", rec.Code)
	}
	var list []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("JSON decode: %v — body: %s", err, rec.Body.String())
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(list))
	}
}

func TestGetTopic_found(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/nmap", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "nmap") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestGetTopic_notFound(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/doesnotexist", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d, want 404", rec.Code)
	}
}

func TestGetTopic_rawParam(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/nmap?raw=1", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d", rec.Code)
	}
	// raw should include the frontmatter delimiter
	if !strings.Contains(rec.Body.String(), "---") {
		t.Fatalf("expected raw content with frontmatter, got: %s", rec.Body.String())
	}
}

func TestGetTopic_JSON(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/nmap?format=json", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d", rec.Code)
	}
	var cs map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &cs); err != nil {
		t.Fatalf("JSON decode: %v", err)
	}
	if cs["topic"] != "nmap" {
		t.Fatalf("topic: got %v", cs["topic"])
	}
}
