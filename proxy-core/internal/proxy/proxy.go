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
	config        Config
	store         store.SessionStore
	httpSrv       *http.Server
	transport     *http.Transport
	authority     *cert.Authority
	fallbackCache *fallbackCache
}

func NewServer(cfg Config, st store.SessionStore, authority *cert.Authority) *Server {
	dialer := &net.Dialer{
		Timeout: cfg.ConnectDialTimeout,
	}

	transport := &http.Transport{
		Proxy:                 nil,
		DialContext:           dialer.DialContext,
		TLSHandshakeTimeout:   cfg.UpstreamTLSHandshakeDelay,
		ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: cfg.UpstreamTLSSkipVerify,
		},
	}

	s := &Server{
		config:        cfg,
		store:         st,
		transport:     transport,
		authority:     authority,
		fallbackCache: newFallbackCache(),
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
	if outReq.Host == "" {
		outReq.Host = outReq.URL.Host
	}
	if scheme == "https" {
		setTLSServerName(outReq)
	}

	resp, err := s.transport.RoundTrip(outReq)
	if err != nil {
		item.EndTime = time.Now()
		item.DurationMS = item.EndTime.Sub(start).Milliseconds()
		item.Error = describeUpstreamError(err)
		_ = s.store.Save(item)
		http.Error(w, item.Error, statusCodeForUpstreamError(err))
		return
	}
	defer resp.Body.Close()

	respBody, respBodyReader, respSize, bodyErr := cloneBody(resp.Body, s.config.MaxBodyBytes)
	if bodyErr != nil && item.Error == "" {
		item.Error = fmt.Sprintf("response body read error: %v", bodyErr)
	}

	for key, values := range sanitizeResponseHeaders(resp.Header, int64(len(respBody))) {
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

	if shouldBypassMITM(host, s.config.MITMBypassHosts) || s.shouldTemporarilyBypassMITM(host) || s.authority == nil {
		if err := s.respondConnectEstablished(clientConn, &item, start); err != nil {
			return
		}
		s.runConnectPassthrough(clientConn, r, &item, start)
		return
	}

	if err := s.runConnectMITM(clientConn, r, &item, start); err != nil {
		s.markMITMFailure(host)
		item.Error = err.Error()
		item.TLSIntercepted = false
		item.EndTime = time.Now()
		item.DurationMS = item.EndTime.Sub(start).Milliseconds()
		_ = s.store.Save(item)
	}
}

func (s *Server) runConnectMITM(clientConn net.Conn, connectReq *http.Request, tunnelItem *session.Session, start time.Time) error {
	if err := s.respondConnectEstablished(clientConn, tunnelItem, start); err != nil {
		return err
	}

	tlsCert, err := s.authority.TLSCertificateForHost(connectReq.Host)
	if err != nil {
		return fmt.Errorf("leaf certificate generation failed: %w", err)
	}

	serverTLSConn := tls.Server(clientConn, &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		MinVersion:   tls.VersionTLS12,
	})
	defer serverTLSConn.Close()

	if err := serverTLSConn.Handshake(); err != nil {
		return fmt.Errorf("client TLS handshake failed: %w", err)
	}

	tunnelItem.TLSIntercepted = true
	tunnelItem.EndTime = time.Now()
	tunnelItem.DurationMS = tunnelItem.EndTime.Sub(start).Milliseconds()
	_ = s.store.Save(*tunnelItem)

	reader := bufio.NewReader(serverTLSConn)

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
		setTLSServerName(req)

		writer := newBufferedMITMResponseWriter(serverTLSConn)
		s.forwardCapturedHTTP(writer, req.WithContext(context.Background()), "https", true)
		if err := writer.Flush(); err != nil {
			return fmt.Errorf("write intercepted response failed: %w", err)
		}

		if shouldCloseConnection(req.Header, writer.Header()) {
			return nil
		}
	}
}

func (s *Server) runConnectPassthrough(clientConn net.Conn, connectReq *http.Request, item *session.Session, start time.Time) {
	upstreamConn, err := net.DialTimeout("tcp", connectReq.Host, s.config.ConnectDialTimeout)
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

func (s *Server) respondConnectEstablished(clientConn net.Conn, item *session.Session, start time.Time) error {
	_, err := clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	if err != nil {
		item.Error = fmt.Sprintf("connect acknowledgement failed: %v", err)
		item.EndTime = time.Now()
		item.DurationMS = item.EndTime.Sub(start).Milliseconds()
		_ = s.store.Save(*item)
		return err
	}
	item.StatusCode = http.StatusOK
	return nil
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

func sanitizeResponseHeaders(h http.Header, bodyLength int64) http.Header {
	cleaned := make(http.Header, len(h)+1)
	for key, values := range h {
		canonical := http.CanonicalHeaderKey(key)
		if canonical == "Transfer-Encoding" {
			continue
		}
		for _, value := range values {
			cleaned.Add(canonical, value)
		}
	}
	cleaned.Set("Content-Length", strconv.FormatInt(bodyLength, 10))
	return cleaned
}

func setTLSServerName(r *http.Request) {
	if r == nil {
		return
	}
	if r.Host == "" && r.URL != nil {
		r.Host = r.URL.Host
	}
	if r.URL != nil && r.URL.Host == "" {
		r.URL.Host = r.Host
	}
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

type bufferedMITMResponseWriter struct {
	conn        net.Conn
	header      http.Header
	statusCode  int
	wroteHeader bool
	body        bytes.Buffer
}

func newBufferedMITMResponseWriter(conn net.Conn) *bufferedMITMResponseWriter {
	return &bufferedMITMResponseWriter{
		conn:   conn,
		header: make(http.Header),
	}
}

func (w *bufferedMITMResponseWriter) Header() http.Header {
	return w.header
}

func (w *bufferedMITMResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.statusCode = statusCode
}

func (w *bufferedMITMResponseWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.body.Write(data)
}

func (w *bufferedMITMResponseWriter) Flush() error {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	headers := sanitizeResponseHeaders(w.header, int64(w.body.Len()))
	statusText := http.StatusText(w.statusCode)
	if statusText == "" {
		statusText = "Status"
	}
	if _, err := fmt.Fprintf(w.conn, "HTTP/1.1 %d %s\r\n", w.statusCode, statusText); err != nil {
		return err
	}
	for key, values := range headers {
		for _, value := range values {
			if _, err := fmt.Fprintf(w.conn, "%s: %s\r\n", key, value); err != nil {
				return err
			}
		}
	}
	if _, err := fmt.Fprintf(w.conn, "\r\n"); err != nil {
		return err
	}
	_, err := io.Copy(w.conn, bytes.NewReader(w.body.Bytes()))
	return err
}

func shouldBypassMITM(host string, bypassHosts []string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if host == "" {
		return false
	}
	for _, pattern := range bypassHosts {
		pattern = strings.ToLower(strings.TrimSpace(pattern))
		if pattern == "" {
			continue
		}
		if host == pattern || strings.HasSuffix(host, "."+pattern) {
			return true
		}
	}
	return false
}

func shouldCloseConnection(requestHeaders, responseHeaders http.Header) bool {
	return headerHasToken(requestHeaders, "Connection", "close") || headerHasToken(responseHeaders, "Connection", "close")
}

func headerHasToken(h http.Header, key, token string) bool {
	for _, value := range h.Values(key) {
		parts := strings.Split(value, ",")
		for _, part := range parts {
			if strings.EqualFold(strings.TrimSpace(part), token) {
				return true
			}
		}
	}
	return false
}

func isExpectedDisconnect(err error) bool {
	return errors.Is(err, io.EOF) ||
		errors.Is(err, net.ErrClosed) ||
		strings.Contains(strings.ToLower(err.Error()), "closed network connection")
}
