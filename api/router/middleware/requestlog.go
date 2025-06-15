package middleware

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"production-go-api-template/pkg/constants"
	"production-go-api-template/pkg/contextkeys"
	"production-go-api-template/pkg/logger"
	"strings"
	"time"
)

const (
	maxLogBodySize = 1024 * 10
)

type logEntry struct {
	RequestID      string              `json:"request_id,omitempty"`
	ReceivedTime   time.Time           `json:"received_time"`
	RequestMethod  string              `json:"request_method"`
	RequestURL     string              `json:"request_url"`
	RequestHeader  map[string][]string `json:"request_header"`
	RequestBody    string              `json:"request_body"`
	UserAgent      string              `json:"user_agent"`
	Referer        string              `json:"referer"`
	Proto          string              `json:"proto"`
	RemoteIP       string              `json:"remote_ip"`
	ServerIP       string              `json:"server_ip,omitempty"`
	Status         int                 `json:"status"`
	ResponseHeader map[string][]string `json:"response_header"`
	ResponseBody   string              `json:"response_body"`
	Latency        time.Duration       `json:"latency"`
}

type responseStats struct {
	w       http.ResponseWriter
	code    int
	bodyBuf bytes.Buffer
}

func RequestLog(l *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			start := time.Now()

			bodyBytes := readAndRestoreBody(r, l)
			reqHeader := sanitizeHeaders(r.Header)
			loggedReqBody := truncateBody(bodyBytes)

			le := newLogEntry(r, reqHeader, loggedReqBody, start)
			le.ServerIP = getServerIP(r)

			w2 := &responseStats{w: w}
			next.ServeHTTP(w2, r)

			finalizeEntry(le, w2)
			l.Info().Fields(entryFields(le)).Msg("http_request")
		})
	}
}

func readAndRestoreBody(r *http.Request, l *logger.Logger) []byte {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		l.Error().Err(err).Msg("failed to read request body")
		bodyBytes = nil
	}
	_ = r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	return bodyBytes
}

func sanitizeHeaders(header http.Header) map[string][]string {
	reqHeader := make(map[string][]string, len(header))
	for k, vals := range header {
		reqHeader[k] = append([]string(nil), vals...)
	}
	for _, h := range []string{"Authorization", "X-Signature", "X-Timestamp"} {
		if _, ok := reqHeader[h]; ok {
			reqHeader[h] = []string{"[REDACTED]"}
		}
	}
	return reqHeader
}

func truncateBody(body []byte) string {
	if len(body) <= maxLogBodySize {
		return string(body)
	}
	return string(body[:maxLogBodySize]) + "...(truncated)"
}

func newLogEntry(r *http.Request, header map[string][]string, body string, start time.Time) *logEntry {
	return &logEntry{
		RequestID:     getRequestIDFromContext(r),
		ReceivedTime:  start,
		RequestMethod: r.Method,
		RequestURL:    r.URL.String(),
		RequestHeader: header,
		RequestBody:   body,
		UserAgent:     r.UserAgent(),
		Referer:       r.Referer(),
		Proto:         r.Proto,
		RemoteIP:      extractClientIP(r),
	}
}

func getRequestIDFromContext(r *http.Request) string {
	if requestID := r.Context().Value(contextkeys.CtxKeyRequestID); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

func getServerIP(r *http.Request) string {
	if addr, ok := r.Context().Value(http.LocalAddrContextKey).(net.Addr); ok {
		host, _, err := net.SplitHostPort(addr.String())
		if err == nil {
			return host
		}
	}
	return ""
}

func finalizeEntry(le *logEntry, stats *responseStats) {
	le.Status = stats.code
	if le.Status == constants.ZeroIndex {
		le.Status = http.StatusOK
	}
	repHeader := make(map[string][]string, len(stats.w.Header()))
	for k, vals := range stats.w.Header() {
		repHeader[k] = append([]string(nil), vals...)
	}
	le.ResponseHeader = repHeader

	fullResp := stats.bodyBuf.String()
	le.ResponseBody = truncateBody([]byte(fullResp))
	le.Latency = time.Since(le.ReceivedTime)
}

func entryFields(le *logEntry) map[string]any {
	return map[string]any{
		"request_id":      le.RequestID,
		"received_time":   le.ReceivedTime,
		"method":          le.RequestMethod,
		"url":             le.RequestURL,
		"request_header":  le.RequestHeader,
		"request_body":    le.RequestBody,
		"agent":           le.UserAgent,
		"referer":         le.Referer,
		"proto":           le.Proto,
		"remote_ip":       le.RemoteIP,
		"server_ip":       le.ServerIP,
		"status":          le.Status,
		"response_header": le.ResponseHeader,
		"response_body":   le.ResponseBody,
		"latency":         le.Latency,
	}
}

func extractClientIP(r *http.Request) string {
	if h := r.Header.Get("X-Forwarded-For"); h != "" {
		parts := strings.Split(h, ",")
		return strings.TrimSpace(parts[constants.ZeroIndex])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (r *responseStats) Header() http.Header {
	return r.w.Header()
}

func (r *responseStats) WriteHeader(statusCode int) {
	if r.code != constants.ZeroIndex {
		return
	}
	r.w.WriteHeader(statusCode)
	r.code = statusCode
}

func (r *responseStats) Write(p []byte) (int, error) {
	if r.code == constants.ZeroIndex {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.w.Write(p)
	r.bodyBuf.Write(p[:n])
	return n, err
}
