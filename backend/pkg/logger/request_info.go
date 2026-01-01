package logger

import (
	"net/http"
	"time"
)

// RequestInfo はリクエストに関する情報を保持する構造体です
type RequestInfo struct {
	Method     string            `json:"method"`
	Path       string            `json:"path"`
	Query      string            `json:"query,omitempty"`
	UserAgent  string            `json:"user_agent,omitempty"`
	RemoteAddr string            `json:"remote_addr"`
	Protocol   string            `json:"protocol"`
	Headers    map[string]string `json:"headers,omitempty"`
	StartTime  time.Time         `json:"start_time"`
}

// NewRequestInfo はHTTPリクエストから RequestInfo を生成します
func NewRequestInfo(r *http.Request) *RequestInfo {
	headers := make(map[string]string)
	// 重要なヘッダーのみを選択して記録
	for _, header := range []string{"Accept", "Content-Type", "X-Forwarded-For", "X-Real-IP"} {
		if value := r.Header.Get(header); value != "" {
			headers[header] = value
		}
	}

	return &RequestInfo{
		Method:     r.Method,
		Path:       r.URL.Path,
		Query:      r.URL.RawQuery,
		UserAgent:  r.UserAgent(),
		RemoteAddr: r.RemoteAddr,
		Protocol:   r.Proto,
		Headers:    headers,
		StartTime:  time.Now(),
	}
}