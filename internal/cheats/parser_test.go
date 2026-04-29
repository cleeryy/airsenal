package cheats

import (
	"testing"
)

func TestParse_plainText(t *testing.T) {
	raw := "# curl\nTransfer data with URLs.\n\ncurl -X GET https://example.com\n"
	cs := Parse("curl", raw)

	if cs.Topic != "curl" {
		t.Fatalf("topic: got %q, want %q", cs.Topic, "curl")
	}
	if cs.Description != "curl" {
		t.Fatalf("description: got %q, want %q", cs.Description, "curl")
	}
	if cs.Content != raw {
		t.Fatalf("content mismatch")
	}
	if cs.Raw != raw {
		t.Fatalf("raw mismatch")
	}
}

func TestParse_frontmatter(t *testing.T) {
	raw := "---\ndescription: Transfer data with URLs\ntags: [network, http]\n---\n\n## Usage\ncurl <url>\n"
	cs := Parse("curl", raw)

	if cs.Description != "Transfer data with URLs" {
		t.Fatalf("description: got %q", cs.Description)
	}
	if len(cs.Tags) != 2 || cs.Tags[0] != "network" || cs.Tags[1] != "http" {
		t.Fatalf("tags: got %v", cs.Tags)
	}
	if cs.Content == "" {
		t.Fatal("body should not be empty")
	}
	if cs.Raw != raw {
		t.Fatal("raw should preserve the original content")
	}
}

func TestParse_malformedFrontmatter(t *testing.T) {
	raw := "---\ndescription: no closing delimiter\n"
	cs := Parse("test", raw)
	if cs.Content != raw {
		t.Fatalf("expected raw fallback when frontmatter is malformed")
	}
}

func TestParse_noTagsInitializesSlice(t *testing.T) {
	raw := "---\ndescription: no tags\n---\nbody"
	cs := Parse("test", raw)
	if cs.Tags == nil {
		t.Fatalf("tags should be empty slice, not nil")
	}
}
