package validator

import (
	"context"
	"errors"
)

func ExtractAndValidateContext[T any](ctx context.Context, key any) (T, error) {
	var zero T
	val := ctx.Value(key)
	if cast, ok := val.(T); ok {
		return cast, nil
	}
	return zero, errors.New("missing or invalid context value")
}
