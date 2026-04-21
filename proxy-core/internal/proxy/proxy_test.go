package proxy

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"netclaw/proxy-core/internal/store"
)

func TestHandleHTTPProxyRequestCapturesSession(t *testing.T) {
	t.Parallel()

	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("hello from target"))
	}))
	defer target.Close()

	st := store.NewMemoryStore()
	s := NewServer(DefaultConfig(), st, nil)

	req, err := http.NewRequest(http.MethodGet, target.URL+"/hello?x=1", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	req.Host = req.URL.Host
	req.RemoteAddr = "127.0.0.1:54321"

	rr := httptest.NewRecorder()
	s.handle(rr, req)

	resp := rr.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(body) != "hello from target" {
		t.Fatalf("body = %q, want %q", string(body), "hello from target")
	}

	items, err := st.List(store.ListOptions{Host: req.URL.Hostname()})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("List() len = %d, want 1", len(items))
	}
	if items[0].Host != req.URL.Hostname() {
		t.Fatalf("captured host = %q, want %q", items[0].Host, req.URL.Hostname())
	}
}

func TestConnectRequestGetsTunnelEstablishedInsteadOfMuxRedirect(t *testing.T) {
	t.Parallel()

	upstream, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen upstream error = %v", err)
	}
	defer upstream.Close()

	payload := "hello-through-tunnel"
	go func() {
		conn, err := upstream.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		buf := make([]byte, len(payload))
		_, _ = io.ReadFull(conn, buf)
		_, _ = conn.Write([]byte(payload))
	}()

	proxyListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen proxy error = %v", err)
	}
	proxyAddr := proxyListener.Addr().String()
	_ = proxyListener.Close()

	cfg := DefaultConfig()
	cfg.ListenAddress = proxyAddr
	st := store.NewMemoryStore()
	s := NewServer(cfg, st, nil)

	go func() {
		_ = s.Start()
	}()
	defer func() { _ = s.Close() }()

	time.Sleep(100 * time.Millisecond)

	conn, err := net.Dial("tcp", proxyAddr)
	if err != nil {
		t.Fatalf("net.Dial proxy error = %v", err)
	}
	defer conn.Close()

	_, err = fmt.Fprintf(conn, "CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", upstream.Addr().String(), upstream.Addr().String())
	if err != nil {
		t.Fatalf("write CONNECT error = %v", err)
	}

	reader := bufio.NewReader(conn)
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	statusLine, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read status line error = %v", err)
	}
	if !strings.Contains(statusLine, "200 Connection Established") {
		t.Fatalf("status line = %q, want 200 Connection Established", statusLine)
	}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("read header line error = %v", err)
		}
		if line == "\r\n" {
			break
		}
	}

	_, err = conn.Write([]byte(payload))
	if err != nil {
		t.Fatalf("write tunneled payload error = %v", err)
	}
	response := make([]byte, len(payload))
	_, err = io.ReadFull(reader, response)
	if err != nil {
		t.Fatalf("read tunneled payload error = %v", err)
	}
	if string(response) != payload {
		t.Fatalf("response payload = %q, want %q", string(response), payload)
	}

	_ = conn.Close()
	time.Sleep(100 * time.Millisecond)

	items, err := st.List(store.ListOptions{Method: http.MethodConnect})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("connect session count = %d, want 1", len(items))
	}
	full, err := st.Get(items[0].ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if full.CaptureMode != "connect-passthrough" {
		t.Fatalf("CaptureMode = %q", full.CaptureMode)
	}
	if full.TunnelBytesUp == 0 || full.TunnelBytesDown == 0 {
		t.Fatalf("tunnel bytes = %d/%d, want non-zero", full.TunnelBytesUp, full.TunnelBytesDown)
	}
}
