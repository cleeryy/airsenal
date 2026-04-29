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
