package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"netclaw/proxy-core/internal/cert"
	"netclaw/proxy-core/internal/store"
)

type Server struct {
	store     store.SessionStore
	authority *cert.Authority
}

func NewServer(st store.SessionStore, authority *cert.Authority) *Server {
	return &Server{store: st, authority: authority}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.health)
	mux.HandleFunc("/api/sessions", s.listSessions)
	mux.HandleFunc("/api/sessions/", s.getSession)
	mux.HandleFunc("/api/certificate-authority", s.certificateAuthorityInfo)
	return mux
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) listSessions(w http.ResponseWriter, _ *http.Request) {
	items, err := s.store.List()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) getSession(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/sessions/")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing id"})
		return
	}
	item, err := s.store.Get(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) certificateAuthorityInfo(w http.ResponseWriter, _ *http.Request) {
	if s.authority == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "certificate authority unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, s.authority.Info())
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
