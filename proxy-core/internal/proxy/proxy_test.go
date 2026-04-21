package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

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
