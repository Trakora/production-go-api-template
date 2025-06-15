package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"production-go-api-template/pkg/contextkeys"
)

const (
	RequestIDHeader = "X-Request-ID"
	requestIDLength = 16
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)

		if requestID == "" {
			requestID = generateRequestID()
		}

		w.Header().Set(RequestIDHeader, requestID)

		ctx := context.WithValue(r.Context(), contextkeys.CtxKeyRequestID, requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateRequestID() string {
	bytes := make([]byte, requestIDLength)
	if _, err := rand.Read(bytes); err != nil {
		return hex.EncodeToString([]byte("fallback"))
	}
	return hex.EncodeToString(bytes)
}
