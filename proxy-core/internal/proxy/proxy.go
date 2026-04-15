package proxy

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"netclaw/proxy-core/internal/cert"
	"netclaw/proxy-core/internal/session"
	"netclaw/proxy-core/internal/store"
)

type Server struct {
	config    Config
	store     store.SessionStore
	httpSrv   *http.Server
	transport *http.Transport
	authority *cert.Authority
}

func NewServer(cfg Config, st store.SessionStore, authority *cert.Authority) *Server {
	transport := &http.Transport{
		Proxy: nil,
	}

	s := &Server{
		config:    cfg,
		store:     st,
		transport: transport,
		authority: authority,
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
	s.forwardCapturedHTTP(w, r, requestScheme(r), false)
}

func (s *Server) forwardCapturedHTTP(w http.ResponseWriter, r *http.Request, scheme string, tlsIntercepted bool) {
	start := time.Now()
	item := session.Session{
		ID:             uuid.NewString(),
		StartTime:      start,
		Scheme:         scheme,
		Method:         r.Method,
		Host:           requestHost(r),
		Port:           requestPortWithScheme(r, scheme),
		Path:           requestPath(r),
		URL:            requestURLWithScheme(r, scheme),
		ClientAddress:  r.RemoteAddr,
		RequestHeaders: flattenHeaders(r.Header),
		TLSIntercepted: tlsIntercepted,
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
		outReq.URL.Scheme = scheme
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

	_, err = clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	if err != nil {
		item.Error = fmt.Sprintf("connect acknowledgement failed: %v", err)
		item.EndTime = time.Now()
		item.DurationMS = item.EndTime.Sub(start).Milliseconds()
		_ = s.store.Save(item)
		return
	}

	item.StatusCode = http.StatusOK

	if s.authority == nil {
		s.runConnectPassthrough(clientConn, r, &item, start)
		return
	}

	if err := s.runConnectMITM(clientConn, r, &item, start); err != nil {
		item.Error = err.Error()
		item.TLSIntercepted = false
		_ = s.store.Save(item)
	}
}

func (s *Server) runConnectMITM(clientConn net.Conn, connectReq *http.Request, tunnelItem *session.Session, start time.Time) error {
	tlsCert, err := s.authority.TLSCertificateForHost(connectReq.Host)
	if err != nil {
		tunnelItem.EndTime = time.Now()
		tunnelItem.DurationMS = tunnelItem.EndTime.Sub(start).Milliseconds()
		return fmt.Errorf("leaf certificate generation failed: %w", err)
	}

	serverTLSConn := tls.Server(clientConn, &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		MinVersion:   tls.VersionTLS12,
	})
	defer serverTLSConn.Close()

	if err := serverTLSConn.Handshake(); err != nil {
		tunnelItem.EndTime = time.Now()
		tunnelItem.DurationMS = tunnelItem.EndTime.Sub(start).Milliseconds()
		return fmt.Errorf("client TLS handshake failed: %w", err)
	}

	tunnelItem.TLSIntercepted = true
	tunnelItem.EndTime = time.Now()
	tunnelItem.DurationMS = tunnelItem.EndTime.Sub(start).Milliseconds()
	_ = s.store.Save(*tunnelItem)

	reader := bufio.NewReader(serverTLSConn)
	writer := &mitmResponseWriter{conn: serverTLSConn, header: make(http.Header)}

	for {
		req, err := http.ReadRequest(reader)
		if err != nil {
			if isExpectedDisconnect(err) {
				return nil
			}
			return fmt.Errorf("read intercepted request failed: %w", err)
		}

		req.RemoteAddr = connectReq.RemoteAddr
		req.URL.Scheme = "https"
		if req.URL.Host == "" {
			req.URL.Host = connectReq.Host
		}
		if req.Host == "" {
			req.Host = connectReq.Host
		}

		writer.reset()
		s.forwardCapturedHTTP(writer, req.WithContext(context.Background()), "https", true)
		if err := writer.flush(); err != nil {
			return fmt.Errorf("write intercepted response failed: %w", err)
		}
	}
}

func (s *Server) runConnectPassthrough(clientConn net.Conn, connectReq *http.Request, item *session.Session, start time.Time) {
	upstreamConn, err := net.DialTimeout("tcp", connectReq.Host, 10*time.Second)
	if err != nil {
		item.Error = fmt.Sprintf("upstream dial failed: %v", err)
		item.EndTime = time.Now()
		item.DurationMS = item.EndTime.Sub(start).Milliseconds()
		_, _ = clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		_ = s.store.Save(*item)
		return
	}
	defer upstreamConn.Close()

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
	_ = s.store.Save(*item)
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
	return requestPortWithScheme(r, requestScheme(r))
}

func requestPortWithScheme(r *http.Request, scheme string) int {
	if r.URL != nil && r.URL.Port() != "" {
		if p, err := strconv.Atoi(r.URL.Port()); err == nil {
			return p
		}
	}
	_, port := splitHostPort(r.Host)
	if port != 0 {
		return port
	}
	if scheme == "https" {
		return 443
	}
	return 80
}

func requestHost(r *http.Request) string {
	if r.URL != nil && r.URL.Hostname() != "" {
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
	return requestURLWithScheme(r, requestScheme(r))
}

func requestURLWithScheme(r *http.Request, scheme string) string {
	if r.URL != nil && r.URL.IsAbs() && r.URL.String() != "" {
		return r.URL.String()
	}
	host := r.Host
	if r.URL != nil && r.URL.Host != "" {
		host = r.URL.Host
	}
	return scheme + "://" + host + requestPath(r)
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

type mitmResponseWriter struct {
	conn        net.Conn
	header      http.Header
	statusCode  int
	wroteHeader bool
}

func (w *mitmResponseWriter) Header() http.Header {
	return w.header
}

func (w *mitmResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.statusCode = statusCode
}

func (w *mitmResponseWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.conn.Write(data)
}

func (w *mitmResponseWriter) reset() {
	w.header = make(http.Header)
	w.statusCode = 0
	w.wroteHeader = false
}

func (w *mitmResponseWriter) flush() error {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	statusText := http.StatusText(w.statusCode)
	if statusText == "" {
		statusText = "Status"
	}
	if _, err := fmt.Fprintf(w.conn, "HTTP/1.1 %d %s\r\n", w.statusCode, statusText); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w.conn, "Connection: keep-alive\r\n"); err != nil {
		return err
	}
	for key, values := range w.header {
		for _, value := range values {
			if _, err := fmt.Fprintf(w.conn, "%s: %s\r\n", key, value); err != nil {
				return err
			}
		}
	}
	_, err := fmt.Fprintf(w.conn, "\r\n")
	return err
}

func isExpectedDisconnect(err error) bool {
	return errors.Is(err, io.EOF) ||
		errors.Is(err, net.ErrClosed) ||
		strings.Contains(strings.ToLower(err.Error()), "closed network connection")
}
