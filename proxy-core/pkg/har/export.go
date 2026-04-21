package har

import (
	"encoding/base64"
	"net/url"
	"sort"
	"strings"
	"time"

	"netclaw/proxy-core/internal/session"
)

type Document struct {
	Log Log `json:"log"`
}

type Log struct {
	Version string  `json:"version"`
	Creator Creator `json:"creator"`
	Entries []Entry `json:"entries"`
}

type Creator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Entry struct {
	StartedDateTime time.Time `json:"startedDateTime"`
	Time            float64   `json:"time"`
	Request         Request   `json:"request"`
	Response        Response  `json:"response"`
	Timings         Timings   `json:"timings"`
	ServerIPAddress string    `json:"serverIPAddress,omitempty"`
}

type Request struct {
	Method      string      `json:"method"`
	URL         string      `json:"url"`
	HTTPVersion string      `json:"httpVersion"`
	Headers     []NameValue `json:"headers"`
	QueryString []NameValue `json:"queryString"`
	HeadersSize int64       `json:"headersSize"`
	BodySize    int64       `json:"bodySize"`
	PostData    *PostData   `json:"postData,omitempty"`
}

type Response struct {
	Status      int         `json:"status"`
	StatusText  string      `json:"statusText"`
	HTTPVersion string      `json:"httpVersion"`
	Headers     []NameValue `json:"headers"`
	Content     Content     `json:"content"`
	RedirectURL string      `json:"redirectURL"`
	HeadersSize int64       `json:"headersSize"`
	BodySize    int64       `json:"bodySize"`
}

type Content struct {
	Size        int64  `json:"size"`
	MimeType    string `json:"mimeType"`
	Text        string `json:"text,omitempty"`
	Encoding    string `json:"encoding,omitempty"`
	Compression int64  `json:"compression,omitempty"`
}

type PostData struct {
	MimeType string `json:"mimeType"`
	Text     string `json:"text"`
}

type NameValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Timings struct {
	Send    float64 `json:"send"`
	Wait    float64 `json:"wait"`
	Receive float64 `json:"receive"`
}

func Export(sessions []session.Session) Document {
	entries := make([]Entry, 0, len(sessions))
	for _, item := range sessions {
		if item.Method == "CONNECT" {
			continue
		}
		entries = append(entries, entryFromSession(item))
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].StartedDateTime.Before(entries[j].StartedDateTime)
	})

	return Document{
		Log: Log{
			Version: "1.2",
			Creator: Creator{Name: "NetClaw", Version: "0.1.0"},
			Entries: entries,
		},
	}
}

func entryFromSession(item session.Session) Entry {
	reqURL := item.URL
	parsedURL, _ := url.Parse(reqURL)
	queryString := []NameValue{}
	if parsedURL != nil {
		for key, values := range parsedURL.Query() {
			for _, value := range values {
				queryString = append(queryString, NameValue{Name: key, Value: value})
			}
		}
	}

	request := Request{
		Method:      item.Method,
		URL:         reqURL,
		HTTPVersion: "HTTP/1.1",
		Headers:     mapToNameValues(item.RequestHeaders),
		QueryString: queryString,
		HeadersSize: -1,
		BodySize:    item.RequestSize,
	}
	if len(item.RequestBody) > 0 {
		request.PostData = &PostData{
			MimeType: headerValue(item.RequestHeaders, "Content-Type"),
			Text:     renderBody(item.RequestBody),
		}
	}

	response := Response{
		Status:      item.StatusCode,
		StatusText:  statusText(item.StatusCode, item.Error),
		HTTPVersion: "HTTP/1.1",
		Headers:     mapToNameValues(item.ResponseHeaders),
		RedirectURL: headerValue(item.ResponseHeaders, "Location"),
		HeadersSize: -1,
		BodySize:    item.ResponseSize,
		Content: Content{
			Size:     item.ResponseSize,
			MimeType: item.ContentType,
			Text:     renderBody(item.ResponseBody),
		},
	}
	if !isTextLike(item.ContentType) && len(item.ResponseBody) > 0 {
		response.Content.Encoding = "base64"
	}

	return Entry{
		StartedDateTime: item.StartTime,
		Time:            float64(item.DurationMS),
		Request:         request,
		Response:        response,
		Timings: Timings{
			Send:    0,
			Wait:    float64(item.DurationMS),
			Receive: 0,
		},
		ServerIPAddress: item.Host,
	}
}

func mapToNameValues(headers map[string]string) []NameValue {
	if len(headers) == 0 {
		return []NameValue{}
	}
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]NameValue, 0, len(keys))
	for _, key := range keys {
		result = append(result, NameValue{Name: key, Value: headers[key]})
	}
	return result
}

func headerValue(headers map[string]string, key string) string {
	for headerKey, value := range headers {
		if strings.EqualFold(headerKey, key) {
			return value
		}
	}
	return ""
}

func renderBody(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	if text := string(data); isLikelyUTF8(text) {
		return text
	}
	return base64.StdEncoding.EncodeToString(data)
}

func isTextLike(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.HasPrefix(contentType, "text/") ||
		strings.Contains(contentType, "json") ||
		strings.Contains(contentType, "xml") ||
		strings.Contains(contentType, "javascript") ||
		strings.Contains(contentType, "x-www-form-urlencoded")
}

func isLikelyUTF8(value string) bool {
	return strings.ToValidUTF8(value, "") == value
}

func statusText(statusCode int, fallback string) string {
	if fallback != "" && statusCode == 0 {
		return fallback
	}
	return ""
}
