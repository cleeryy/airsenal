package cheats

import (
	"strings"
	"testing"
)

func TestParseArsenal(t *testing.T) {
	raw := "# nmap\n\n% network, scan, recon\n\n#plateform/linux #target/remote #cat/RECON\n\n## nmap - classic scan\n#tag/scan\nThis is a description.\nIt can span multiple lines.\n\n```\nnmap -sC -sV <ip>\n```\n\n## nmap - udp scan\n```\nnmap -sU <ip>\n```\n\n$ var: ls /tmp\n= const: /tmp/wordlist.txt\n"

	cs := ParseArsenal("nmap", "Scan", raw)

	if cs.Topic != "nmap" {
		t.Errorf("Topic = %q, want nmap", cs.Topic)
	}
	if cs.Category != "Scan" {
		t.Errorf("Category = %q, want Scan", cs.Category)
	}
	if cs.Description != "nmap" {
		t.Errorf("Description = %q, want nmap", cs.Description)
	}
	if len(cs.Tags) != 3 {
		t.Errorf("Tags len = %d, want 3", len(cs.Tags))
	} else {
		if cs.Tags[0] != "network" || cs.Tags[1] != "scan" || cs.Tags[2] != "recon" {
			t.Errorf("Tags = %v", cs.Tags)
		}
	}

	expectedContent := strings.TrimSpace(`
## nmap - classic scan
This is a description.
It can span multiple lines.


` + "```\n" + `nmap -sC -sV <ip>` + "\n```\n" + `
## nmap - udp scan

` + "```\n" + `nmap -sU <ip>` + "\n```")

	if cs.Content != expectedContent {
		t.Errorf("Content mismatch.\nGot:\n%s\n\nWant:\n%s", cs.Content, expectedContent)
	}
}

func TestIsCommandTagLine(t *testing.T) {
	cases := []struct {
		line string
		want bool
	}{
		{"#plateform/linux #target/remote", true},
		{"#cat/RECON", true},
		{"#", false},
		{"## section", false},
		{"# title", false},
		{" #tag", false},
	}

	for _, tc := range cases {
		got := isCommandTagLine(tc.line)
		if got != tc.want {
			t.Errorf("isCommandTagLine(%q) = %v, want %v", tc.line, got, tc.want)
		}
	}
}
