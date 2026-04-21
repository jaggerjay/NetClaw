package har

import (
	"testing"
	"time"

	"netclaw/proxy-core/internal/session"
)

func TestExportSkipsConnectAndBuildsEntries(t *testing.T) {
	t.Parallel()

	doc := Export([]session.Session{
		{
			ID:          "connect-1",
			StartTime:   time.Now().UTC().Add(-2 * time.Second),
			Method:      "CONNECT",
			Host:        "example.com",
			URL:         "https://example.com:443",
			CaptureMode: "connect-mitm",
		},
		{
			ID:              "get-1",
			StartTime:       time.Now().UTC().Add(-1 * time.Second),
			Method:          "GET",
			Host:            "example.com",
			URL:             "https://example.com/hello?x=1",
			StatusCode:      200,
			DurationMS:      42,
			RequestHeaders:  map[string]string{"Accept": "*/*"},
			ResponseHeaders: map[string]string{"Content-Type": "text/plain"},
			ResponseBody:    []byte("hello"),
			ResponseSize:    5,
			ContentType:     "text/plain",
		},
	})

	if got := len(doc.Log.Entries); got != 1 {
		t.Fatalf("len(entries) = %d, want 1", got)
	}
	entry := doc.Log.Entries[0]
	if entry.Request.Method != "GET" {
		t.Fatalf("Request.Method = %q", entry.Request.Method)
	}
	if entry.Response.Status != 200 {
		t.Fatalf("Response.Status = %d", entry.Response.Status)
	}
	if entry.Response.Content.Text != "hello" {
		t.Fatalf("Response.Content.Text = %q", entry.Response.Content.Text)
	}
	if len(entry.Request.QueryString) != 1 || entry.Request.QueryString[0].Name != "x" {
		t.Fatalf("unexpected query string: %#v", entry.Request.QueryString)
	}
}
