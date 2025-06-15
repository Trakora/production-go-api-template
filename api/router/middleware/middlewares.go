package middleware

import (
	"net/http"
	"production-go-api-template/pkg/constants"
)

type Middleware func(http.Handler) http.Handler

func CreateStack(xs ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(xs) - constants.FirstIndex; i >= constants.ZeroIndex; i-- {
			x := xs[i]
			next = x(next)
		}

		return next
	}
}
