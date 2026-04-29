package api

import (
	"net/http"

	"github.com/cleeryy/airsenal/internal/cheats"
)

// NewRouter wires all HTTP routes and returns the root handler.
func NewRouter(store *cheats.Store) http.Handler {
	h := newHandler(store)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("GET /~search", h.search)
	mux.HandleFunc("GET /", h.list)
	mux.HandleFunc("GET /{topic}", h.getTopic)

	return mux
}
