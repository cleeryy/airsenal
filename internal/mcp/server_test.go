package mcp

import (
	"bytes"
	"encoding/json"
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
	if _, err := s.Load(); err != nil {
		t.Fatal(err)
	}
	return s
}

func runMessages(t *testing.T, msgs ...string) []map[string]interface{} {
	t.Helper()
	srv := NewServer(newTestStore(t), "test-version")
	input := strings.Join(msgs, "\n") + "\n"
	var out bytes.Buffer
	if err := srv.run(strings.NewReader(input), &out); err != nil {
		t.Fatalf("run: %v", err)
	}
	var results []map[string]interface{}
	for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
		if line == "" {
			continue
		}
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("decode response line %q: %v", line, err)
		}
		results = append(results, m)
	}
	return results
}

func TestMCP_initialize(t *testing.T) {
	responses := runMessages(t, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"test","version":"0.1"}}}`)
	if len(responses) != 1 {
		t.Fatalf("expected 1 response, got %d", len(responses))
	}
	res := responses[0]
	if res["error"] != nil {
		t.Fatalf("unexpected error: %v", res["error"])
	}
	result, ok := res["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("result not an object: %v", res["result"])
	}
	if result["protocolVersion"] != protocolVersion {
		t.Fatalf("protocolVersion: got %v", result["protocolVersion"])
	}
}

func TestMCP_initialized_noResponse(t *testing.T) {
	responses := runMessages(t, `{"jsonrpc":"2.0","method":"initialized"}`)
	if len(responses) != 0 {
		t.Fatalf("notifications should not produce a response, got %d responses", len(responses))
	}
}

func TestMCP_toolsList(t *testing.T) {
	responses := runMessages(t, `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
	if len(responses) != 1 {
		t.Fatalf("expected 1 response")
	}
	result := responses[0]["result"].(map[string]interface{})
	toolsRaw := result["tools"].([]interface{})
	if len(toolsRaw) != 6 {
		t.Fatalf("expected 6 tools, got %d", len(toolsRaw))
	}
}

func TestMCP_randomCheatsheet(t *testing.T) {
	responses := runMessages(t,
		`{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"random_cheatsheet","arguments":{}}}`,
	)
	if len(responses) != 1 {
		t.Fatalf("expected 1 response")
	}
	result := responses[0]["result"].(map[string]interface{})
	content := result["content"].([]interface{})
	text := content[0].(map[string]interface{})["text"].(string)

	var cs map[string]interface{}
	if err := json.Unmarshal([]byte(text), &cs); err != nil {
		t.Fatalf("failed to decode random response: %v", err)
	}
	if cs["topic"] != "nmap" {
		t.Fatalf("expected nmap topic, got %v", cs["topic"])
	}
	if cs["random"] != true {
		t.Fatalf("expected random: true, got %v", cs["random"])
	}
}

func TestMCP_getCheatsheet(t *testing.T) {
	responses := runMessages(t,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_cheatsheet","arguments":{"topic":"nmap"}}}`,
	)
	if len(responses) != 1 {
		t.Fatalf("expected 1 response")
	}
	result := responses[0]["result"].(map[string]interface{})
	content := result["content"].([]interface{})
	text := content[0].(map[string]interface{})["text"].(string)
	if !strings.Contains(text, "nmap") {
		t.Fatalf("expected nmap in response, got: %s", text)
	}
}

func TestMCP_searchCheatsheets(t *testing.T) {
	responses := runMessages(t,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"search_cheatsheets","arguments":{"query":"network"}}}`,
	)
	if len(responses) != 1 {
		t.Fatalf("expected 1 response")
	}
	result := responses[0]["result"].(map[string]interface{})
	content := result["content"].([]interface{})
	text := content[0].(map[string]interface{})["text"].(string)
	if !strings.Contains(text, "nmap") {
		t.Fatalf("expected nmap in search response, got: %s", text)
	}
	if !strings.Contains(text, "[tag]") {
		t.Fatalf("expected [tag] match reason in response, got: %s", text)
	}
}

func TestMCP_searchCheatsheets_noResults(t *testing.T) {
	responses := runMessages(t,
		`{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"search_cheatsheets","arguments":{"query":"zzznomatch"}}}`,
	)
	if len(responses) != 1 {
		t.Fatalf("expected 1 response")
	}
	result := responses[0]["result"].(map[string]interface{})
	content := result["content"].([]interface{})
	text := content[0].(map[string]interface{})["text"].(string)
	if !strings.Contains(text, "No cheatsheets matched") {
		t.Fatalf("expected no-match message, got: %s", text)
	}
}

func TestMCP_unknownMethod(t *testing.T) {
	responses := runMessages(t, `{"jsonrpc":"2.0","id":99,"method":"nonexistent"}`)
	if len(responses) != 1 {
		t.Fatalf("expected 1 error response")
	}
	if responses[0]["error"] == nil {
		t.Fatal("expected error for unknown method")
	}
}
