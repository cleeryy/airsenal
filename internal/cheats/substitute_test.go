package cheats

import (
	"strings"
	"testing"
)

func TestExtractVariables_basic(t *testing.T) {
	content := "nmap <target>\nnmap -p <port> <target>"
	vars := ExtractVariables(content)
	if len(vars) != 2 {
		t.Fatalf("expected 2 vars, got %d: %v", len(vars), vars)
	}
	if vars[0] != "port" || vars[1] != "target" {
		t.Fatalf("unexpected vars: %v", vars)
	}
}

func TestExtractVariables_none(t *testing.T) {
	vars := ExtractVariables("no placeholders here")
	if len(vars) != 0 {
		t.Fatalf("expected 0 vars, got %v", vars)
	}
}

func TestExtractVariables_sorted(t *testing.T) {
	vars := ExtractVariables("<zoo> <apple> <mango>")
	if vars[0] != "apple" || vars[1] != "mango" || vars[2] != "zoo" {
		t.Fatalf("not sorted: %v", vars)
	}
}

func TestSubstitute_noDecoration(t *testing.T) {
	result := Substitute("nmap <target>", map[string]string{"target": "10.10.10.10"}, false)
	if result != "nmap 10.10.10.10" {
		t.Fatalf("unexpected: %q", result)
	}
}

func TestSubstitute_decoration(t *testing.T) {
	result := Substitute("nmap <target>", map[string]string{"target": "10.10.10.10"}, true)
	if !strings.Contains(result, "\x1b[1m10.10.10.10\x1b[0m") {
		t.Fatalf("expected ANSI bold wrapping, got: %q", result)
	}
	if strings.Contains(result, "<target>") {
		t.Fatalf("placeholder was not replaced: %q", result)
	}
}

func TestSubstitute_multipleOccurrences(t *testing.T) {
	result := Substitute("<target>\n<target>", map[string]string{"target": "host"}, false)
	if result != "host\nhost" {
		t.Fatalf("unexpected: %q", result)
	}
}

func TestSubstitute_partialProvision(t *testing.T) {
	result := Substitute("nmap -p <port> <target>", map[string]string{"target": "host"}, false)
	if !strings.Contains(result, "<port>") {
		t.Fatalf("unprovided var should remain: %q", result)
	}
	if !strings.Contains(result, "host") {
		t.Fatalf("provided var should be substituted: %q", result)
	}
}

func TestSubstitute_noVars(t *testing.T) {
	content := "nmap <target>"
	result := Substitute(content, map[string]string{}, false)
	if result != content {
		t.Fatalf("empty vars should be a no-op: %q", result)
	}
}
