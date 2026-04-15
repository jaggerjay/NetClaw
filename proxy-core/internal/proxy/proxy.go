package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
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
		s.handleConnect(w, r)
		return
	}

	start := time.Now()
	item := session.Session{
		ID:             uuid.NewString(),
		StartTime:      start,
		Scheme:         requestScheme(r),
		Method:         r.Method,
		Host:           requestHost(r),
		Port:           requestPort(r),
		Path:           requestPath(r),
		URL:            requestURL(r),
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

func (s *Server) handleConnect(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	host, port := splitHostPort(r.Host)
	item := session.Session{
		ID:            uuid.NewString(),
		StartTime:     start,
		EndTime:       start,
		Scheme:        "https",
		Method:        http.MethodConnect,
		Host:          host,
		Port:          port,
		Path:          "",
		URL:           "https://" + r.Host,
		ClientAddress: r.RemoteAddr,
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		item.Error = "response writer does not support hijacking"
		_ = s.store.Save(item)
		http.Error(w, item.Error, http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		item.Error = fmt.Sprintf("client hijack failed: %v", err)
		item.EndTime = time.Now()
		item.DurationMS = item.EndTime.Sub(start).Milliseconds()
		_ = s.store.Save(item)
		return
	}
	defer clientConn.Close()

	upstreamConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		item.Error = fmt.Sprintf("upstream dial failed: %v", err)
		item.EndTime = time.Now()
		item.DurationMS = item.EndTime.Sub(start).Milliseconds()
		_, _ = clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		_ = s.store.Save(item)
		return
	}
	defer upstreamConn.Close()

	_, err = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	if err != nil {
		item.Error = fmt.Sprintf("connect acknowledgement failed: %v", err)
		item.EndTime = time.Now()
		item.DurationMS = item.EndTime.Sub(start).Milliseconds()
		_ = s.store.Save(item)
		return
	}

	item.StatusCode = http.StatusOK

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, _ = io.Copy(upstreamConn, clientConn)
		if tcpConn, ok := upstreamConn.(*net.TCPConn); ok {
			_ = tcpConn.CloseWrite()
		}
	}()

	go func() {
		defer wg.Done()
		_, _ = io.Copy(clientConn, upstreamConn)
		if tcpConn, ok := clientConn.(*net.TCPConn); ok {
			_ = tcpConn.CloseWrite()
		}
	}()

	wg.Wait()
	item.EndTime = time.Now()
	item.DurationMS = item.EndTime.Sub(start).Milliseconds()
	item.TLSIntercepted = false
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
	_, port := splitHostPort(r.Host)
	if port != 0 {
		return port
	}
	if requestScheme(r) == "https" {
		return 443
	}
	return 80
}

func requestHost(r *http.Request) string {
	if r.URL.Hostname() != "" {
		return r.URL.Hostname()
	}
	host, _ := splitHostPort(r.Host)
	return host
}

func requestPath(r *http.Request) string {
	if r.URL == nil {
		return ""
	}
	if r.URL.RequestURI() != "" {
		return r.URL.RequestURI()
	}
	return r.URL.Path
}

func requestURL(r *http.Request) string {
	if r.URL != nil && r.URL.String() != "" {
		return r.URL.String()
	}
	return requestScheme(r) + "://" + r.Host + requestPath(r)
}

func splitHostPort(value string) (string, int) {
	if value == "" {
		return "", 0
	}
	if host, portStr, err := net.SplitHostPort(value); err == nil {
		if port, convErr := strconv.Atoi(portStr); convErr == nil {
			return host, port
		}
		return host, 0
	}
	return value, 443
}
