package contextkeys

import (
	"context"
)

func GetRequestID(ctx context.Context) string {
	if requestID := ctx.Value(CtxKeyRequestID); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}
