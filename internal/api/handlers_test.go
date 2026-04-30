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
	content := "---\ndescription: Port scanner\ncategory: Scan\ntags: [network]\n---\n\n## Usage\nnmap <target>\n"
	if err := os.WriteFile(filepath.Join(dir, "nmap.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	s := cheats.NewStore(dir)
	if _, err := s.Load(); err != nil {
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

func TestGetTopic_substitution(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/nmap?target=10.10.10.10", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "10.10.10.10") {
		t.Fatalf("substituted value missing: %s", body)
	}
	// Check only the content part (before the Variables: hint line)
	contentPart, _, _ := strings.Cut(body, "\nVariables:")
	if strings.Contains(contentPart, "<target>") {
		t.Fatalf("placeholder should have been replaced in content: %s", contentPart)
	}
	// ANSI bold decoration present
	if !strings.Contains(body, "\x1b[1m10.10.10.10\x1b[0m") {
		t.Fatalf("expected ANSI bold wrapping in plain-text output: %q", body)
	}
}

func TestGetTopic_substitutionRaw(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/nmap?raw=1&target=10.10.10.10", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "10.10.10.10") {
		t.Fatalf("substituted value missing in raw output: %s", body)
	}
	// No ANSI in raw mode
	if strings.Contains(body, "\x1b[") {
		t.Fatalf("raw output should not contain ANSI codes: %q", body)
	}
}

func TestGetTopic_unprovidedVarRemains(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/nmap", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "<target>") {
		t.Fatalf("<target> should remain when no vars given: %s", body)
	}
}

func TestGetTopic_helperLine(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/nmap", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, "Variables:") {
		t.Fatalf("expected variables hint line: %s", body)
	}
	if !strings.Contains(body, "<target>") {
		t.Fatalf("expected <target> in hint line: %s", body)
	}
}

func TestGetTopic_helperLineNotInJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/nmap?format=json", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	body := rec.Body.String()
	if strings.Contains(body, "Variables:") {
		t.Fatalf("variables hint should not appear in JSON output: %s", body)
	}
}

func TestGetTopic_helperLineNotInRaw(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/nmap?raw=1", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	body := rec.Body.String()
	if strings.Contains(body, "Variables:") {
		t.Fatalf("variables hint should not appear in raw output: %s", body)
	}
}

func TestGetTopic_reservedParamsNotSubstituted(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/nmap?format=json&raw=1", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	// format=json wins; body must be valid JSON and not treat "format"/"raw" as vars
	var cs map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &cs); err != nil {
		t.Fatalf("JSON decode: %v — body: %s", err, rec.Body.String())
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

func TestGetTopic_terminalANSI(t *testing.T) {
	for _, ua := range []string{
		"curl/8.1.2",
		"Wget/1.21.3",
		"HTTPie/3.2.2",
		"Go-http-client/1.1",
	} {
		t.Run(ua, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/nmap", nil)
			req.Header.Set("User-Agent", ua)
			NewRouter(newTestStore(t)).ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status: got %d", rec.Code)
			}
			body := rec.Body.String()
			if !strings.Contains(body, "\x1b[") {
				t.Fatalf("terminal client %q should receive ANSI codes, got: %q", ua, body)
			}
		})
	}
}

func TestGetTopic_nonTerminalNoANSI(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/nmap", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	body := rec.Body.String()
	if strings.Contains(body, "\x1b[33m") {
		t.Fatalf("browser UA must not receive ANSI section-header codes: %q", body)
	}
}

func TestListCategories_plainText(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/~categories", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Available categories (1)") {
		t.Fatalf("expected 1 category, got: %s", body)
	}
	if !strings.Contains(body, "Scan") {
		t.Fatalf("expected Scan in list, got: %s", body)
	}
}

func TestListCategories_JSON(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/~categories?format=json", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d", rec.Code)
	}
	var list []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("JSON decode: %v — body: %s", err, rec.Body.String())
	}
	if len(list) != 1 || list[0]["category"] != "Scan" {
		t.Fatalf("expected 1 entry for Scan, got %v", list)
	}
}

func TestListByCategory_found(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/~cat/scan", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "nmap") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestListByCategory_notFound(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/~cat/doesnotexist", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d, want 404", rec.Code)
	}
}

func TestListTags_plainText(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/~tags", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Available tags (1)") {
		t.Fatalf("expected 1 tag, got: %s", body)
	}
	if !strings.Contains(body, "network") {
		t.Fatalf("expected network in list, got: %s", body)
	}
}

func TestListByTag_found(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/~tag/network", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "nmap") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestListByTag_notFound(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/~tag/doesnotexist", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status: got %d, want 404", rec.Code)
	}
}

func TestSearch_tagMatch(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/~search?q=network", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "nmap") {
		t.Fatalf("expected nmap in results, got: %s", body)
	}
	if !strings.Contains(body, "[tags:") {
		t.Fatalf("expected [tags: ...] annotation for tag match, got: %s", body)
	}
}

func TestSearch_topicMatch(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/~search?q=nmap", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "nmap") {
		t.Fatalf("expected nmap in results, got: %s", body)
	}
	// topic matches carry no annotation — self-evident from the topic name
	if strings.Contains(body, "[tags:") || strings.Contains(body, "match]") {
		t.Fatalf("topic match should have no annotation, got: %s", body)
	}
}

func TestSearch_usageFooter(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/~search?q=nmap", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	if !strings.Contains(rec.Body.String(), "Usage:") {
		t.Fatalf("expected Usage footer in plain-text output, got: %s", rec.Body.String())
	}
}

func TestSearch_noResults(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/~search?q=zzznomatch", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "No results") {
		t.Fatalf("expected 'No results', got: %s", body)
	}
}

func TestSearch_missingQuery(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/~search", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want 400", rec.Code)
	}
}

func TestSearch_JSON(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/~search?q=nmap&format=json", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rec.Code)
	}
	var results []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &results); err != nil {
		t.Fatalf("JSON decode: %v — body: %s", err, rec.Body.String())
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0]["topic"] != "nmap" {
		t.Fatalf("topic: got %v", results[0]["topic"])
	}
	if results[0]["match_reason"] != "topic" {
		t.Fatalf("match_reason: got %v, want \"topic\"", results[0]["match_reason"])
	}
}

func TestSearch_JSONNoResults(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/~search?q=zzznomatch&format=json", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rec.Code)
	}
	var results []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &results); err != nil {
		t.Fatalf("JSON decode: %v — body: %s", err, rec.Body.String())
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestSearch_limit(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/~search?q=nmap&limit=1", nil)
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rec.Code)
	}
}

func TestRandom(t *testing.T) {
	s := newTestStore(t)
	router := NewRouter(s)

	t.Run("basic", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/~random", nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status: got %d", rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "Random cheat: nmap") {
			t.Fatalf("expected header, got: %s", body)
		}
		if !strings.Contains(body, "nmap <target>") {
			t.Fatalf("expected content, got: %s", body)
		}
	})

	t.Run("category found", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/~random?cat=Scan", nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status: got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "nmap") {
			t.Fatalf("expected nmap, got: %s", rec.Body.String())
		}
	})

	t.Run("category not found", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/~random?cat=doesnotexist", nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status: got %d, want 404", rec.Code)
		}
	})

	t.Run("tag found", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/~random?tag=network", nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status: got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "nmap") {
			t.Fatalf("expected nmap, got: %s", rec.Body.String())
		}
	})

	t.Run("JSON", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/~random?format=json", nil)
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status: got %d", rec.Code)
		}
		var resp map[string]interface{}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if resp["random"] != true {
			t.Fatalf("expected random: true, got %v", resp["random"])
		}
		if resp["topic"] != "nmap" {
			t.Fatalf("expected topic: nmap, got %v", resp["topic"])
		}
	})
}

func TestGetTopic_rawSkipsANSI(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/nmap?raw=1", nil)
	req.Header.Set("User-Agent", "curl/8.1.2")
	NewRouter(newTestStore(t)).ServeHTTP(rec, req)

	body := rec.Body.String()
	if strings.Contains(body, "\x1b[33m") {
		t.Fatalf("raw=1 must bypass ANSI rendering even for curl: %q", body)
	}
}
