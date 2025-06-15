package validator

import (
	"encoding/json"
	"net/http"
)

type Validator interface {
	Validate() error
}

func DecodeAndValidate[T Validator](r *http.Request, v T) error {
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return err
	}
	return v.Validate()
}
