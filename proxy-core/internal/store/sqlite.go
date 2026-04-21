package store

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"netclaw/proxy-core/internal/session"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	if strings.TrimSpace(dbPath) == "" {
		return nil, fmt.Errorf("sqlite db path is required")
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	store := &SQLiteStore{db: db}
	if err := store.init(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *SQLiteStore) init() error {
	statements := []string{
		`PRAGMA journal_mode = WAL;`,
		`PRAGMA busy_timeout = 5000;`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			start_time TEXT NOT NULL,
			end_time TEXT NOT NULL,
			scheme TEXT NOT NULL,
			method TEXT NOT NULL,
			host TEXT NOT NULL,
			port INTEGER NOT NULL,
			path TEXT NOT NULL,
			url TEXT NOT NULL,
			status_code INTEGER NOT NULL,
			duration_ms INTEGER NOT NULL,
			client_address TEXT NOT NULL,
			request_headers_json TEXT NOT NULL,
			response_headers_json TEXT NOT NULL,
			request_body_base64 TEXT NOT NULL,
			response_body_base64 TEXT NOT NULL,
			request_size INTEGER NOT NULL,
			response_size INTEGER NOT NULL,
			content_type TEXT NOT NULL,
			error_text TEXT NOT NULL,
			tls_intercepted INTEGER NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_start_time ON sessions(start_time DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_host ON sessions(host);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_method ON sessions(method);`,
	}

	for _, statement := range statements {
		if _, err := s.db.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteStore) Save(item session.Session) error {
	requestHeadersJSON, err := json.Marshal(item.RequestHeaders)
	if err != nil {
		return err
	}
	responseHeadersJSON, err := json.Marshal(item.ResponseHeaders)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		INSERT INTO sessions (
			id, start_time, end_time, scheme, method, host, port, path, url,
			status_code, duration_ms, client_address,
			request_headers_json, response_headers_json,
			request_body_base64, response_body_base64,
			request_size, response_size, content_type, error_text, tls_intercepted
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			start_time=excluded.start_time,
			end_time=excluded.end_time,
			scheme=excluded.scheme,
			method=excluded.method,
			host=excluded.host,
			port=excluded.port,
			path=excluded.path,
			url=excluded.url,
			status_code=excluded.status_code,
			duration_ms=excluded.duration_ms,
			client_address=excluded.client_address,
			request_headers_json=excluded.request_headers_json,
			response_headers_json=excluded.response_headers_json,
			request_body_base64=excluded.request_body_base64,
			response_body_base64=excluded.response_body_base64,
			request_size=excluded.request_size,
			response_size=excluded.response_size,
			content_type=excluded.content_type,
			error_text=excluded.error_text,
			tls_intercepted=excluded.tls_intercepted
	`,
		item.ID,
		item.StartTime.UTC().Format(time.RFC3339Nano),
		item.EndTime.UTC().Format(time.RFC3339Nano),
		item.Scheme,
		item.Method,
		item.Host,
		item.Port,
		item.Path,
		item.URL,
		item.StatusCode,
		item.DurationMS,
		item.ClientAddress,
		string(requestHeadersJSON),
		string(responseHeadersJSON),
		base64.StdEncoding.EncodeToString(item.RequestBody),
		base64.StdEncoding.EncodeToString(item.ResponseBody),
		item.RequestSize,
		item.ResponseSize,
		item.ContentType,
		item.Error,
		boolToInt(item.TLSIntercepted),
	)
	return err
}

func (s *SQLiteStore) List(options ListOptions) ([]session.Summary, error) {
	rows, err := s.db.Query(`
		SELECT id, start_time, method, host, url, status_code, duration_ms, content_type, response_size, error_text
		FROM sessions
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]session.Summary, 0)
	for rows.Next() {
		var (
			item      session.Summary
			startTime string
			errorText string
		)
		if err := rows.Scan(
			&item.ID,
			&startTime,
			&item.Method,
			&item.Host,
			&item.URL,
			&item.StatusCode,
			&item.DurationMS,
			&item.ContentType,
			&item.ResponseSize,
			&errorText,
		); err != nil {
			return nil, err
		}
		parsed, err := time.Parse(time.RFC3339Nano, startTime)
		if err != nil {
			return nil, err
		}
		item.StartTime = parsed
		item.HasError = errorText != ""
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	filtered := make([]session.Summary, 0, len(result))
	for _, summary := range result {
		full, err := s.Get(summary.ID)
		if err != nil {
			return nil, err
		}
		if !matchesListOptions(*full, options) {
			continue
		}
		filtered = append(filtered, summary)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].StartTime.After(filtered[j].StartTime)
	})
	if options.Limit > 0 && len(filtered) > options.Limit {
		filtered = filtered[:options.Limit]
	}
	return filtered, nil
}

func (s *SQLiteStore) Get(id string) (*session.Session, error) {
	row := s.db.QueryRow(`
		SELECT id, start_time, end_time, scheme, method, host, port, path, url,
			status_code, duration_ms, client_address,
			request_headers_json, response_headers_json,
			request_body_base64, response_body_base64,
			request_size, response_size, content_type, error_text, tls_intercepted
		FROM sessions
		WHERE id = ?
	`, id)

	var (
		item                session.Session
		startTime           string
		endTime             string
		requestHeadersJSON  string
		responseHeadersJSON string
		requestBodyBase64   string
		responseBodyBase64  string
		tlsIntercepted      int
	)

	err := row.Scan(
		&item.ID,
		&startTime,
		&endTime,
		&item.Scheme,
		&item.Method,
		&item.Host,
		&item.Port,
		&item.Path,
		&item.URL,
		&item.StatusCode,
		&item.DurationMS,
		&item.ClientAddress,
		&requestHeadersJSON,
		&responseHeadersJSON,
		&requestBodyBase64,
		&responseBodyBase64,
		&item.RequestSize,
		&item.ResponseSize,
		&item.ContentType,
		&item.Error,
		&tlsIntercepted,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("session not found")
	}
	if err != nil {
		return nil, err
	}

	item.StartTime, err = time.Parse(time.RFC3339Nano, startTime)
	if err != nil {
		return nil, err
	}
	item.EndTime, err = time.Parse(time.RFC3339Nano, endTime)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(requestHeadersJSON), &item.RequestHeaders); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(responseHeadersJSON), &item.ResponseHeaders); err != nil {
		return nil, err
	}
	if requestBodyBase64 != "" {
		item.RequestBody, err = base64.StdEncoding.DecodeString(requestBodyBase64)
		if err != nil {
			return nil, err
		}
	}
	if responseBodyBase64 != "" {
		item.ResponseBody, err = base64.StdEncoding.DecodeString(responseBodyBase64)
		if err != nil {
			return nil, err
		}
	}
	item.TLSIntercepted = tlsIntercepted != 0
	return &item, nil
}

func (s *SQLiteStore) Clear() error {
	_, err := s.db.Exec(`DELETE FROM sessions`)
	return err
}

func (s *SQLiteStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
