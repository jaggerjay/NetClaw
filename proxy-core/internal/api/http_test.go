package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"netclaw/proxy-core/internal/session"
	"netclaw/proxy-core/internal/store"
)

func TestListSessionsFilters(t *testing.T) {
	t.Parallel()

	st := store.NewMemoryStore()
	seedSession(t, st, session.Session{
		ID:             "1",
		StartTime:      time.Now().UTC().Add(-2 * time.Minute),
		EndTime:        time.Now().UTC().Add(-119 * time.Second),
		Scheme:         "https",
		Method:         "GET",
		Host:           "example.com",
		Path:           "/ok",
		URL:            "https://example.com/ok",
		StatusCode:     200,
		ResponseSize:   10,
		ContentType:    "text/plain",
		TLSIntercepted: true,
	})
	seedSession(t, st, session.Session{
		ID:             "2",
		StartTime:      time.Now().UTC().Add(-1 * time.Minute),
		EndTime:        time.Now().UTC().Add(-59 * time.Second),
		Scheme:         "https",
		Method:         "POST",
		Host:           "api.example.com",
		Path:           "/fail",
		URL:            "https://api.example.com/fail",
		StatusCode:     502,
		ResponseSize:   20,
		ContentType:    "application/json",
		Error:          "upstream timeout",
		TLSIntercepted: false,
	})

	ts := httptest.NewServer(NewServer(st, nil, RuntimeInfo{}).Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/sessions?host=api.example.com&method=POST&has_error=true&tls_intercepted=false&limit=1")
	if err != nil {
		t.Fatalf("GET /api/sessions error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var items []session.Summary
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		t.Fatalf("decode error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(items))
	}
	if items[0].ID != "2" {
		t.Fatalf("items[0].ID = %q, want %q", items[0].ID, "2")
	}
}

func TestRuntimeInfo(t *testing.T) {
	t.Parallel()

	st := store.NewMemoryStore()
	ts := httptest.NewServer(NewServer(st, nil, RuntimeInfo{
		ProxyListenAddress: "127.0.0.1:9090",
		APIListenAddress:   "127.0.0.1:9091",
		DataDir:            ".netclaw-data/dev",
		CertificatePath:    ".netclaw-data/dev/certs/netclaw-root-ca.pem",
	}).Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/runtime-info")
	if err != nil {
		t.Fatalf("GET /api/runtime-info error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var info RuntimeInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		t.Fatalf("decode error = %v", err)
	}
	if info.ProxyListenAddress != "127.0.0.1:9090" {
		t.Fatalf("ProxyListenAddress = %q", info.ProxyListenAddress)
	}
}

func seedSession(t *testing.T, st *store.MemoryStore, item session.Session) {
	t.Helper()
	if err := st.Save(item); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
}
