package render

import (
	"strings"
	"testing"

	"github.com/cleeryy/airsenal/internal/cheats"
)

func mkCS(content string) *cheats.Cheatsheet {
	return &cheats.Cheatsheet{Topic: "test", Content: content}
}

// --- plain-text mode ---

func TestRenderPlain_noVars(t *testing.T) {
	cs := mkCS("## Section\n    cmd\n")
	out := Render(cs, nil, false)
	if !strings.Contains(out, "## Section") {
		t.Fatalf("expected raw markdown preserved, got: %q", out)
	}
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("plain mode must not contain ANSI codes: %q", out)
	}
}

func TestRenderPlain_substitution(t *testing.T) {
	cs := mkCS("    connect <host>\n")
	out := Render(cs, map[string]string{"host": "10.0.0.1"}, false)
	if !strings.Contains(out, "10.0.0.1") {
		t.Fatalf("substituted value missing: %q", out)
	}
	if !strings.Contains(out, "\x1b[1m10.0.0.1\x1b[0m") {
		t.Fatalf("expected ANSI bold wrapping in plain mode: %q", out)
	}
}

func TestRenderPlain_variablesHint(t *testing.T) {
	cs := mkCS("    cmd <host> <port>\n")
	out := Render(cs, nil, false)
	if !strings.Contains(out, "Variables:") {
		t.Fatalf("expected variables hint line: %q", out)
	}
	if !strings.Contains(out, "<host>") || !strings.Contains(out, "<port>") {
		t.Fatalf("expected all placeholders in hint: %q", out)
	}
}

func TestRenderPlain_noHintWhenNoVars(t *testing.T) {
	cs := mkCS("    cmd\n")
	out := Render(cs, nil, false)
	if strings.Contains(out, "Variables:") {
		t.Fatalf("hint should be absent when no placeholders: %q", out)
	}
}

// --- terminal ANSI mode ---

func TestRenderANSI_topicTitle(t *testing.T) {
	cs := mkCS("# nmap\n")
	out := Render(cs, nil, true)
	if !strings.Contains(out, bold+cyan+"nmap"+reset) {
		t.Fatalf("title must be bold-cyan: %q", out)
	}
	if !strings.Contains(out, cyan+"━━━━"+reset) {
		t.Fatalf("separator must follow title: %q", out)
	}
}

func TestRenderANSI_sectionHeader(t *testing.T) {
	cs := mkCS("## Scan Types\n")
	out := Render(cs, nil, true)
	if !strings.Contains(out, bold+yellow+"Scan Types"+reset) {
		t.Fatalf("section header must be bold-yellow: %q", out)
	}
	if strings.Contains(out, "## ") {
		t.Fatalf("raw ## must be stripped: %q", out)
	}
}

func TestRenderANSI_commandLine(t *testing.T) {
	cs := mkCS("    nmap -sS 10.0.0.1\n")
	out := Render(cs, nil, true)
	if !strings.Contains(out, bold+white) {
		t.Fatalf("command must be bold-white: %q", out)
	}
}

func TestRenderANSI_inlineComment(t *testing.T) {
	cs := mkCS("    nmap -sS <target>                 # SYN scan\n")
	out := Render(cs, nil, true)
	if !strings.Contains(out, gray+" # SYN scan"+reset) {
		t.Fatalf("inline comment must be gray: %q", out)
	}
}

func TestRenderANSI_unsubstitutedPlaceholder(t *testing.T) {
	cs := mkCS("    nmap <target>\n")
	out := Render(cs, nil, true)
	if !strings.Contains(out, bold+green+"<target>"+reset) {
		t.Fatalf("unsubstituted placeholder must be bold-green: %q", out)
	}
}

func TestRenderANSI_substitutedVar(t *testing.T) {
	cs := mkCS("    nmap <target>\n")
	out := Render(cs, map[string]string{"target": "192.168.1.1"}, true)
	if !strings.Contains(out, bold+magenta+"192.168.1.1"+reset) {
		t.Fatalf("substituted value must be bold-magenta: %q", out)
	}
	if strings.Contains(out, "<target>") {
		t.Fatalf("placeholder must be replaced: %q", out)
	}
}

func TestRenderANSI_codeFenceStripped(t *testing.T) {
	cs := mkCS("```\nnmap -sV <target>\n```\n")
	out := Render(cs, nil, true)
	if strings.Contains(out, "```") {
		t.Fatalf("backtick fences must be stripped: %q", out)
	}
	if !strings.Contains(out, bold+white) {
		t.Fatalf("fenced code must be rendered as command: %q", out)
	}
}

func TestRenderANSI_horizontalRuleStripped(t *testing.T) {
	cs := mkCS("---\n")
	out := Render(cs, nil, true)
	if strings.Contains(out, "---") {
		t.Fatalf("horizontal rule must be stripped: %q", out)
	}
}

func TestRenderANSI_footer_unsubstituted(t *testing.T) {
	cs := mkCS("    cmd <host>\n")
	out := Render(cs, nil, true)
	if !strings.Contains(out, bold+green+"<host>"+reset) {
		t.Fatalf("footer must show unsubstituted var in bold-green: %q", out)
	}
	if !strings.Contains(out, dim+"Variables:"+reset) {
		t.Fatalf("footer label must be dim: %q", out)
	}
}

func TestRenderANSI_footer_substituted(t *testing.T) {
	cs := mkCS("    cmd <host>\n")
	out := Render(cs, map[string]string{"host": "db.local"}, true)
	if !strings.Contains(out, bold+magenta+"db.local"+reset) {
		t.Fatalf("footer must show substituted value in bold-magenta: %q", out)
	}
}

func TestRenderANSI_noFooterWhenNoVars(t *testing.T) {
	cs := mkCS("    cmd\n")
	out := Render(cs, nil, true)
	if strings.Contains(out, "Variables:") {
		t.Fatalf("footer must be absent when content has no placeholders: %q", out)
	}
}

func TestRenderANSI_separatorMatchesTitleLength(t *testing.T) {
	cs := mkCS("# curl\n")
	out := Render(cs, nil, true)
	if !strings.Contains(out, cyan+"━━━━"+reset) {
		t.Fatalf("separator must be 4 chars for 'curl': %q", out)
	}
}
