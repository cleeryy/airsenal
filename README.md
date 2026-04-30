# airsenal

A lightweight, self-hosted cheatsheet service for the terminal — inspired by [Arsenal](https://github.com/Orange-Cyberdefense/arsenal) by Orange Cyberdefense.

```bash
curl cheat.example.com/nmap
curl cheat.example.com/ssh
curl cheat.example.com/docker
```

---

## Why it exists

[Arsenal](https://github.com/Orange-Cyberdefense/arsenal) is a brilliant quick-reference launcher for penetration testing commands. `airsenal` takes that same spirit — curated, command-focused cheatsheets — and wraps it in an HTTP service you can query from anywhere with `curl`, and exposes an MCP interface so AI assistants can fetch the same content programmatically.

---

## Features

- **Rich ANSI output** for terminal clients (`curl`, `wget`, HTTPie, Go) — bold-cyan titles, bold-yellow section headers, bold-white commands, dim inline comments, colored variable placeholders
- Plain-text fallback for browsers and API clients
- JSON output via `?format=json` or `Accept: application/json`
- **Template variable substitution** — pass `?key=value` query params to fill `<key>` placeholders inline; unsubstituted placeholders are bold-green, substituted values are bold-magenta
- **Full-text search** via `GET /~search?q=<query>` — ranked across topic name, tags, description, and content
- Content-driven: add a `.md` or `.txt` file to the `cheats/` directory to publish a new cheatsheet
- Optional YAML frontmatter for metadata (description, tags)
- MCP stdio server for AI assistant integration (`AIRSENAL_ENABLE_MCP=true`)
- Single statically-linked binary — no runtime dependencies
- Multi-arch Docker image (`linux/amd64`, `linux/arm64`)

---

## Quick start

### Run with Go

```bash
git clone https://github.com/cleeryy/airsenal
cd airsenal
go run ./cmd/airsenal
# or: make run
```

### Run with Docker

```bash
docker run -p 8080:8080 ghcr.io/cleeryy/airsenal:latest
```

### Run with Docker Compose

```bash
docker compose up -d
```

---

## curl usage

```bash
# List all available topics
curl http://localhost:8080/

# Get a cheatsheet (ANSI-formatted automatically for curl)
curl http://localhost:8080/nmap
curl http://localhost:8080/ssh
curl http://localhost:8080/docker

# Get raw file content — no ANSI, includes frontmatter
curl "http://localhost:8080/nmap?raw=1"

# Get structured JSON
curl "http://localhost:8080/nmap?format=json"
curl -H "Accept: application/json" http://localhost:8080/

# Fill in template variables (bold-magenta in ANSI output)
curl "http://localhost:8080/nmap?target=10.10.10.10"
curl "http://localhost:8080/nmap?target=10.10.10.10&port=443"

# Search across all cheatsheets
curl "http://localhost:8080/~search?q=sql"
curl "http://localhost:8080/~search?q=port+scan&format=json"
curl "http://localhost:8080/~search?q=network&limit=5"

# Health check
curl http://localhost:8080/healthz
```

---

## Terminal formatting

When the request comes from a terminal-oriented client (`curl`, `wget`, HTTPie, or the Go HTTP client), airsenal automatically applies ANSI formatting — no flag required:

| Element | Style |
|---|---|
| Topic title (`# heading`) | Bold cyan + `━━━━` separator |
| Section header (`## heading`) | Bold yellow |
| Command lines (4-space or tab indent, code fences) | Bold white |
| Inline comments (` # comment` after a command) | Dim gray |
| Unsubstituted placeholder `<target>` | Bold green |
| Substituted value | Bold magenta |
| Variables footer | Dim |

Markdown syntax (`##`, `---`, backticks) is stripped and replaced with its ANSI equivalent — the raw source is never shown.

Pass `?raw=1` to skip all formatting and get the original file content including frontmatter.

---

## Template variables

Cheatsheet files use `<variable>` placeholders in command examples (e.g. `<target>`, `<port>`, `<domain>`). Pass these as query parameters to get a pre-filled, ready-to-run copy of the cheatsheet:

```bash
curl "http://localhost:8080/nmap?target=10.10.10.10"
curl "http://localhost:8080/nmap?target=10.10.10.10&port=443"
```

**How it works:**

- Each `?key=value` pair replaces every occurrence of `<key>` in the output.
- In ANSI mode (terminal clients), substituted values are **bold magenta**; unsubstituted placeholders are **bold green**.
- In plain-text mode (browsers, API clients), substituted values are wrapped in ANSI bold only.
- Variables that are not provided are left as-is (e.g. `<port>` stays if `port` is not given).
- A hint line at the bottom of every plain-text response lists all variables available in that cheatsheet:
  ```
  Variables: <port>  <target>
  ```
- `?raw=1` combined with variables still performs substitution but without any decoration.
- `?format=json` returns the unmodified cheatsheet struct — no substitution is applied.
- The reserved params `raw` and `format` are never treated as variable names.

---

## Search

`GET /~search?q=<query>` finds cheatsheets by keyword across all topics. The `~` prefix marks meta-routes that can never conflict with topic names.

```bash
curl "http://localhost:8080/~search?q=sql"
curl "http://localhost:8080/~search?q=port+scan"
curl "http://localhost:8080/~search?q=network&format=json"
curl "http://localhost:8080/~search?q=ssh&limit=5"
```

**Ranking** — results are ordered by where the match was found:

| Priority | Match location |
|----------|---------------|
| 1 (highest) | Topic name |
| 2 | Tags |
| 3 | Description |
| 4 | Content (first 512 chars) |

Within each rank, results are sorted alphabetically by topic name.

**Parameters:**

| Parameter | Default | Description |
|-----------|---------|-------------|
| `q` | *(required)* | Case-insensitive search query |
| `limit` | `20` | Maximum number of results to return |
| `format` | plain text | Set to `json` for a JSON array response |

**Match annotations** — plain-text results show why each item matched:

| Match reason | Annotation |
|---|---|
| Topic name | *(none — self-evident)* |
| Tag | `[tags: tag1, tag2]` |
| Description | `[description match]` |
| Content | `[content match]` |

Plain-text response example:

```
Search results for "sql" (3 matches):

sqlmap SQL injection testing tool
ldapsearch Query LDAP directories [content match]
curl HTTP client [content match]

Usage: curl cheat.example.com/<topic>
```

**JSON output** returns an array of objects. Each object has the same fields as `GET /<topic>?format=json`, plus a `match_reason` field indicating where the query matched:

| `match_reason` value | Meaning |
|---|---|
| `"topic"` | Query matched the topic name |
| `"tag"` | Query matched one of the tags |
| `"description"` | Query matched the description |
| `"content"` | Query matched the cheatsheet body |

---

## Configuration

All configuration is via environment variables:

| Variable              | Default    | Description                                    |
|-----------------------|------------|------------------------------------------------|
| `AIRSENAL_PORT`       | `8080`     | HTTP port to listen on                         |
| `AIRSENAL_CHEATS_DIR` | `./cheats` | Directory containing cheatsheet files          |
| `ARSENAL_CHEATS_DIR`  | `""`       | Directory containing Arsenal cheatsheets       |
| `AIRSENAL_ENABLE_MCP` | `false`    | When `true`, also run the MCP stdio server     |

---

## Arsenal cheats

`airsenal` natively supports parsing the massive cheatsheet library from the upstream [Orange-Cyberdefense/arsenal](https://github.com/Orange-Cyberdefense/arsenal) project.

1. Clone `airsenal` with its submodules:
   ```bash
   git clone --recursive https://github.com/cleeryy/airsenal
   ```
   *(If you already cloned it without submodules, run: `git submodule update --init --recursive`)*
2. Set the `ARSENAL_CHEATS_DIR` environment variable to point to the data directory:
   ```bash
   export ARSENAL_CHEATS_DIR="./arsenal/arsenal/data/cheats"
   ```
3. Start the server. You'll see a log line confirming the loaded arsenal topics.

To pull the latest cheatsheets from the upstream repository, run:
```bash
make update-cheats
```

## Adding custom cheatsheets

Drop a `.md` or `.txt` file into the `cheats/` directory. The filename (without extension) becomes the topic name.

**Minimal example — `cheats/mytool.txt`:**

```
My tool quick reference.

mytool --help
mytool scan <target>
```

**With frontmatter metadata — `cheats/mytool.md`:**

```markdown
---
description: My tool does amazing things
tags: [network, recon]
---

# mytool

## Basic usage
    mytool scan <target>
    mytool scan --verbose <target>
```

Use `<variable>` placeholders in your commands (e.g. `<target>`, `<port>`) — they will be detected automatically and users can substitute them via query parameters (see [Template variables](#template-variables)).

Restart the server (or remount the volume) to pick up new files.

> **Docker tip:** mount your custom cheatsheets directory as a volume:
> ```bash
> docker run -p 8080:8080 -v /my/cheats:/cheats:ro ghcr.io/cleeryy/airsenal:latest
> ```

---

## MCP server

`airsenal` implements the [Model Context Protocol](https://spec.modelcontextprotocol.io/) over stdio, so AI clients (Claude Desktop, Cursor, etc.) can query cheatsheets directly.

### Enable

```bash
AIRSENAL_ENABLE_MCP=true ./airsenal
```

When `AIRSENAL_ENABLE_MCP=true`, the HTTP server starts in a background goroutine and the MCP server claims stdin/stdout for the JSON-RPC protocol stream. All log output goes to stderr.

### Available MCP tools

| Tool                  | Description                                                   |
|-----------------------|---------------------------------------------------------------|
| `list_cheatsheets`    | List all available topics with descriptions                   |
| `get_cheatsheet`      | Get content and metadata for a topic (`topic: string`)        |
| `search_cheatsheets`  | Full-text search across topics, tags, and content (`query: string`) |

### Claude Desktop configuration

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "airsenal": {
      "command": "/path/to/airsenal",
      "env": {
        "AIRSENAL_ENABLE_MCP": "true",
        "AIRSENAL_CHEATS_DIR": "/path/to/cheats"
      }
    }
  }
}
```

---

## Development

### Prerequisites

- Go 1.22+
- Docker (optional, for container builds)

### Commands

```bash
make build       # compile binary to bin/airsenal
make run         # go run ./cmd/airsenal
make test        # go test ./...
make lint        # go vet ./...
make docker-build  # docker build -t airsenal:latest .
make clean       # remove build artifacts
```

### Project structure

```
airsenal/
├── cmd/airsenal/         # entry point
├── internal/
│   ├── api/              # HTTP handlers and router
│   ├── cheats/           # cheatsheet model, parser, and store
│   ├── config/           # environment-based configuration
│   ├── mcp/              # MCP stdio server
│   └── render/           # ANSI terminal renderer
├── cheats/               # bundled cheatsheet content
└── .github/workflows/    # CI/CD pipelines
```

---

## Releases

Pre-built binaries are published on every version tag via GitHub Actions:

| Platform       | File                              |
|----------------|-----------------------------------|
| Linux amd64    | `airsenal-linux-amd64.tar.gz`     |
| Linux arm64    | `airsenal-linux-arm64.tar.gz`     |
| macOS amd64    | `airsenal-darwin-amd64.tar.gz`    |
| macOS arm64    | `airsenal-darwin-arm64.tar.gz`    |
| Windows amd64  | `airsenal-windows-amd64.zip`      |

Docker images are pushed to `ghcr.io/cleeryy/airsenal` on every push to `main` and on version tags.

---

## Bundled cheatsheets

| Topic        | Description                                        |
|--------------|----------------------------------------------------|
| `nmap`       | Network exploration and port scanning              |
| `curl`       | HTTP client and data transfer                      |
| `ssh`        | Secure Shell — remote login, tunneling, key mgmt  |
| `git`        | Version control workflow                           |
| `docker`     | Container management                               |
| `tmux`       | Terminal multiplexer                               |
| `ffuf`       | Fast web fuzzer for content and parameter discovery |
| `sqlmap`     | Automated SQL injection detection and exploitation |
| `ldapsearch` | LDAP directory enumeration                         |
| `dig`        | DNS lookup and zone inspection                     |

---

## License

GPL-3.0 — see [LICENSE](LICENSE).
