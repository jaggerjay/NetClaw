package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"netclaw/proxy-core/internal/cert"
	"netclaw/proxy-core/internal/store"
)

type RuntimeInfo struct {
	ProxyListenAddress string `json:"proxyListenAddress"`
	APIListenAddress   string `json:"apiListenAddress"`
	DataDir            string `json:"dataDir"`
	CertificatePath    string `json:"certificatePath"`
}

type Server struct {
	store       store.SessionStore
	authority   *cert.Authority
	runtimeInfo RuntimeInfo
}

func NewServer(st store.SessionStore, authority *cert.Authority, runtimeInfo RuntimeInfo) *Server {
	return &Server{store: st, authority: authority, runtimeInfo: runtimeInfo}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.health)
	mux.HandleFunc("/api/sessions", s.listSessions)
	mux.HandleFunc("/api/sessions/", s.getSession)
	mux.HandleFunc("/api/certificate-authority", s.certificateAuthorityInfo)
	mux.HandleFunc("/api/runtime-info", s.getRuntimeInfo)
	return mux
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) listSessions(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.List(parseListOptions(r))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func parseListOptions(r *http.Request) store.ListOptions {
	query := r.URL.Query()
	options := store.ListOptions{
		Query:  strings.TrimSpace(query.Get("q")),
		Host:   strings.TrimSpace(query.Get("host")),
		Method: strings.TrimSpace(query.Get("method")),
		Limit:  parsePositiveInt(query.Get("limit")),
	}
	if value, ok := parseOptionalBool(query.Get("has_error")); ok {
		options.HasError = &value
	}
	if value, ok := parseOptionalBool(query.Get("tls_intercepted")); ok {
		options.TLSIntercepted = &value
	}
	return options
}

func parsePositiveInt(value string) int {
	if value == "" {
		return 0
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0
	}
	return parsed
}

func parseOptionalBool(value string) (bool, bool) {
	if value == "" {
		return false, false
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, false
	}
	return parsed, true
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

func (s *Server) getRuntimeInfo(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.runtimeInfo)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
