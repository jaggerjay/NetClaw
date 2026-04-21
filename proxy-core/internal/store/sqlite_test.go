package store

import (
	"path/filepath"
	"testing"
	"time"

	"netclaw/proxy-core/internal/session"
)

func TestSQLiteStoreSaveListGetClear(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	st, err := NewSQLiteStore(filepath.Join(dir, "sessions.sqlite"))
	if err != nil {
		t.Fatalf("NewSQLiteStore() error = %v", err)
	}
	defer func() { _ = st.Close() }()

	item := session.Session{
		ID:                    "session-1",
		StartTime:             time.Now().UTC().Add(-2 * time.Second).Round(0),
		EndTime:               time.Now().UTC().Round(0),
		Scheme:                "https",
		Method:                "GET",
		Host:                  "example.com",
		Port:                  443,
		Path:                  "/hello",
		URL:                   "https://example.com/hello",
		StatusCode:            200,
		DurationMS:            123,
		ClientAddress:         "127.0.0.1:50000",
		RequestHeaders:        map[string]string{"Accept": "*/*"},
		ResponseHeaders:       map[string]string{"Content-Type": "text/plain"},
		RequestBody:           []byte("request"),
		ResponseBody:          []byte("response"),
		RequestSize:           7,
		ResponseSize:          8,
		ContentType:           "text/plain",
		TLSIntercepted:        true,
		CaptureMode:           "https-mitm-request",
		FallbackReason:        "",
		TunnelBytesUp:         111,
		TunnelBytesDown:       222,
		TunnelTargetAddress:   "example.com:443",
		RequestBodyTruncated:  true,
		ResponseBodyTruncated: false,
		RequestBodyEncoding:   "utf8",
		ResponseBodyEncoding:  "utf8",
	}

	if err := st.Save(item); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	items, err := st.List(ListOptions{Query: "example", TLSIntercepted: boolPtr(true)})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("List() len = %d, want 1", len(items))
	}
	if items[0].ID != item.ID {
		t.Fatalf("List()[0].ID = %q, want %q", items[0].ID, item.ID)
	}

	got, err := st.Get(item.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.URL != item.URL {
		t.Fatalf("Get().URL = %q, want %q", got.URL, item.URL)
	}
	if string(got.ResponseBody) != string(item.ResponseBody) {
		t.Fatalf("Get().ResponseBody = %q, want %q", string(got.ResponseBody), string(item.ResponseBody))
	}
	if !got.TLSIntercepted {
		t.Fatalf("Get().TLSIntercepted = false, want true")
	}
	if got.CaptureMode != "https-mitm-request" {
		t.Fatalf("Get().CaptureMode = %q", got.CaptureMode)
	}
	if got.TunnelBytesUp != 111 || got.TunnelBytesDown != 222 {
		t.Fatalf("Get() tunnel bytes = %d/%d", got.TunnelBytesUp, got.TunnelBytesDown)
	}
	if !got.RequestBodyTruncated || got.ResponseBodyTruncated {
		t.Fatalf("unexpected truncation flags: request=%v response=%v", got.RequestBodyTruncated, got.ResponseBodyTruncated)
	}
	if got.RequestBodyEncoding != "utf8" || got.ResponseBodyEncoding != "utf8" {
		t.Fatalf("unexpected encodings: %q / %q", got.RequestBodyEncoding, got.ResponseBodyEncoding)
	}

	if err := st.Clear(); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}
	items, err = st.List(ListOptions{})
	if err != nil {
		t.Fatalf("List() after Clear error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("List() after Clear len = %d, want 0", len(items))
	}
}

func boolPtr(value bool) *bool {
	return &value
}
