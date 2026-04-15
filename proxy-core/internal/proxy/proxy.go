package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"netclaw/proxy-core/internal/session"
	"netclaw/proxy-core/internal/store"
)

type Server struct {
	config    Config
	store     store.SessionStore
	httpSrv   *http.Server
	transport *http.Transport
}

func NewServer(cfg Config, st store.SessionStore) *Server {
	transport := &http.Transport{
		Proxy: nil,
	}

	s := &Server{
		config:    cfg,
		store:     st,
		transport: transport,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handle)

	s.httpSrv = &http.Server{
		Addr:    cfg.ListenAddress,
		Handler: mux,
	}

	return s
}

func (s *Server) Start() error {
	return s.httpSrv.ListenAndServe()
}

func (s *Server) Close() error {
	return s.httpSrv.Close()
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	if strings.ToUpper(r.Method) == http.MethodConnect {
		http.Error(w, "CONNECT MITM not implemented yet in scaffold", http.StatusNotImplemented)
		return
	}

	start := time.Now()
	item := session.Session{
		ID:             uuid.NewString(),
		StartTime:      start,
		Scheme:         requestScheme(r),
		Method:         r.Method,
		Host:           r.URL.Hostname(),
		Port:           requestPort(r),
		Path:           r.URL.RequestURI(),
		URL:            r.URL.String(),
		ClientAddress:  r.RemoteAddr,
		RequestHeaders: flattenHeaders(r.Header),
	}

	if s.config.CaptureBodies {
		bodyBytes, restoredBody, size, err := cloneBody(r.Body, s.config.MaxBodyBytes)
		if err == nil {
			item.RequestBody = bodyBytes
			item.RequestSize = size
			r.Body = restoredBody
		} else {
			item.Error = fmt.Sprintf("request body read error: %v", err)
		}
	}

	outReq := r.Clone(r.Context())
	outReq.RequestURI = ""
	if outReq.URL.Scheme == "" {
		outReq.URL.Scheme = requestScheme(r)
	}
	if outReq.URL.Host == "" {
		outReq.URL.Host = r.Host
	}

	resp, err := s.transport.RoundTrip(outReq)
	if err != nil {
		item.EndTime = time.Now()
		item.DurationMS = item.EndTime.Sub(start).Milliseconds()
		item.Error = err.Error()
		_ = s.store.Save(item)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	respBody, respBodyReader, respSize, bodyErr := cloneBody(resp.Body, s.config.MaxBodyBytes)
	if bodyErr != nil && item.Error == "" {
		item.Error = fmt.Sprintf("response body read error: %v", bodyErr)
	}

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, respBodyReader)

	item.EndTime = time.Now()
	item.DurationMS = item.EndTime.Sub(start).Milliseconds()
	item.StatusCode = resp.StatusCode
	item.ResponseHeaders = flattenHeaders(resp.Header)
	item.ResponseBody = respBody
	item.ResponseSize = respSize
	item.ContentType = resp.Header.Get("Content-Type")
	_ = s.store.Save(item)
}

func cloneBody(rc io.ReadCloser, limit int64) ([]byte, io.ReadCloser, int64, error) {
	if rc == nil {
		return nil, io.NopCloser(bytes.NewReader(nil)), 0, nil
	}
	defer rc.Close()

	limited := io.LimitReader(rc, limit+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, io.NopCloser(bytes.NewReader(nil)), 0, err
	}
	if int64(len(data)) > limit {
		data = data[:limit]
	}
	return data, io.NopCloser(bytes.NewReader(data)), int64(len(data)), nil
}

func flattenHeaders(h http.Header) map[string]string {
	result := make(map[string]string, len(h))
	for k, v := range h {
		result[k] = strings.Join(v, ", ")
	}
	return result
}

func requestScheme(r *http.Request) string {
	if r.URL.Scheme != "" {
		return r.URL.Scheme
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func requestPort(r *http.Request) int {
	if r.URL.Port() != "" {
		if p, err := strconv.Atoi(r.URL.Port()); err == nil {
			return p
		}
	}
	if requestScheme(r) == "https" {
		return 443
	}
	return 80
}
