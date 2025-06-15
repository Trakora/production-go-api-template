package middleware

import (
	"context"
	"net/http"
	"production-go-api-template/config"
	"production-go-api-template/pkg/contextkeys"
	"production-go-api-template/pkg/logger"
)

func InjectDeps(cfg *config.Conf, log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, contextkeys.CtxKeyConfig, cfg)
			ctx = context.WithValue(ctx, contextkeys.CtxKeyLogger, log)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
