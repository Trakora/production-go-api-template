package contextkeys

type ctxKey string

const (
	CtxKeyConfig ctxKey = "config"

	CtxKeyLogger ctxKey = "logger"

	CtxKeyRequestID ctxKey = "request_id"
)
