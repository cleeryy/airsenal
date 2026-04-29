package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cleeryy/airsenal/internal/api"
	"github.com/cleeryy/airsenal/internal/cheats"
	"github.com/cleeryy/airsenal/internal/config"
	"github.com/cleeryy/airsenal/internal/mcp"
)

var version = "dev"

func main() {
	// Always log to stderr so stdout is available for MCP stdio transport.
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ltime | log.Lshortfile)

	cfg := config.Load()

	store := cheats.NewStore(cfg.CheatsDir)
	if err := store.Load(); err != nil {
		log.Printf("warning: %v", err)
	}

	if cfg.EnableMCP {
		// Start HTTP server in background so both interfaces are available.
		go serveHTTP(cfg.Port, store)
		log.Printf("MCP stdio server ready")
		mcpSrv := mcp.NewServer(store, version)
		if err := mcpSrv.RunStdio(); err != nil {
			log.Fatalf("MCP server error: %v", err)
		}
		return
	}

	serveHTTP(cfg.Port, store)
}

func serveHTTP(port string, store *cheats.Store) {
	addr := fmt.Sprintf(":%s", port)
	log.Printf("airsenal %s HTTP server listening on %s", version, addr)
	if err := http.ListenAndServe(addr, api.NewRouter(store)); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}
