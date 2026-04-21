package session

import "time"

type Session struct {
	ID                    string            `json:"id"`
	StartTime             time.Time         `json:"startTime"`
	EndTime               time.Time         `json:"endTime"`
	Scheme                string            `json:"scheme"`
	Method                string            `json:"method"`
	Host                  string            `json:"host"`
	Port                  int               `json:"port"`
	Path                  string            `json:"path"`
	URL                   string            `json:"url"`
	StatusCode            int               `json:"statusCode"`
	DurationMS            int64             `json:"durationMs"`
	ClientAddress         string            `json:"clientAddress"`
	RequestHeaders        map[string]string `json:"requestHeaders"`
	ResponseHeaders       map[string]string `json:"responseHeaders"`
	RequestBody           []byte            `json:"requestBody,omitempty"`
	ResponseBody          []byte            `json:"responseBody,omitempty"`
	RequestSize           int64             `json:"requestSize"`
	ResponseSize          int64             `json:"responseSize"`
	ContentType           string            `json:"contentType"`
	Error                 string            `json:"error,omitempty"`
	TLSIntercepted        bool              `json:"tlsIntercepted"`
	CaptureMode           string            `json:"captureMode"`
	FallbackReason        string            `json:"fallbackReason,omitempty"`
	TunnelBytesUp         int64             `json:"tunnelBytesUp"`
	TunnelBytesDown       int64             `json:"tunnelBytesDown"`
	TunnelTargetAddress   string            `json:"tunnelTargetAddress,omitempty"`
	RequestBodyTruncated  bool              `json:"requestBodyTruncated"`
	ResponseBodyTruncated bool              `json:"responseBodyTruncated"`
	RequestBodyEncoding   string            `json:"requestBodyEncoding,omitempty"`
	ResponseBodyEncoding  string            `json:"responseBodyEncoding,omitempty"`
}

type Summary struct {
	ID              string    `json:"id"`
	StartTime       time.Time `json:"startTime"`
	Method          string    `json:"method"`
	Host            string    `json:"host"`
	URL             string    `json:"url"`
	StatusCode      int       `json:"statusCode"`
	DurationMS      int64     `json:"durationMs"`
	ContentType     string    `json:"contentType"`
	ResponseSize    int64     `json:"responseSize"`
	HasError        bool      `json:"hasError"`
	CaptureMode     string    `json:"captureMode"`
	TLSIntercepted  bool      `json:"tlsIntercepted"`
	TunnelBytesUp   int64     `json:"tunnelBytesUp"`
	TunnelBytesDown int64     `json:"tunnelBytesDown"`
}

func (s Session) ToSummary() Summary {
	return Summary{
		ID:              s.ID,
		StartTime:       s.StartTime,
		Method:          s.Method,
		Host:            s.Host,
		URL:             s.URL,
		StatusCode:      s.StatusCode,
		DurationMS:      s.DurationMS,
		ContentType:     s.ContentType,
		ResponseSize:    s.ResponseSize,
		HasError:        s.Error != "",
		CaptureMode:     s.CaptureMode,
		TLSIntercepted:  s.TLSIntercepted,
		TunnelBytesUp:   s.TunnelBytesUp,
		TunnelBytesDown: s.TunnelBytesDown,
	}
}
