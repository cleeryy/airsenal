package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cleeryy/airsenal/internal/api"
	"github.com/cleeryy/airsenal/internal/cheats"
	"github.com/cleeryy/airsenal/internal/config"
)

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Ltime | log.Lshortfile)

	cfg := config.Load()

	store := cheats.NewStore(cfg.CheatsDir)
	if err := store.Load(); err != nil {
		log.Printf("warning: %v", err)
	}

	router := api.NewRouter(store)
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("airsenal listening on %s (cheats dir: %s)", addr, cfg.CheatsDir)

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
