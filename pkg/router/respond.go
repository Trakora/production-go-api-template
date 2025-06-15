package router

import (
	"encoding/json"
	"net/http"
	"production-go-api-template/pkg/contextkeys"
	"production-go-api-template/pkg/logger"
	"production-go-api-template/pkg/validator"
)

const (
	serverErrorThreshold    = 499
	internalServerErrorCode = 500
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func RespondWithError(
	r *http.Request,
	w http.ResponseWriter,
	code int,
	msg string,
	err error,
) {
	log, ctxErr := validator.ExtractAndValidateContext[*logger.Logger](
		r.Context(), contextkeys.CtxKeyLogger,
	)
	if ctxErr == nil {
		if err != nil {
			log.Errorf("handling error: %v", err)
		}
		if code > serverErrorThreshold {
			log.Warnf("responding with server error %d: %s", code, msg)
		}
	}

	RespondWithJSON(r, w, code, ErrorResponse{Message: msg})
}

func RespondWithJSON(
	r *http.Request,
	w http.ResponseWriter,
	code int,
	payload any,
) {
	log, ctxErr := validator.ExtractAndValidateContext[*logger.Logger](
		r.Context(), contextkeys.CtxKeyLogger,
	)

	data, err := json.Marshal(payload)
	if err != nil {
		if ctxErr == nil {
			log.Errorf("error marshalling JSON: %v", err)
		}
		w.WriteHeader(internalServerErrorCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if _, err := w.Write(data); err != nil {
		if ctxErr == nil {
			log.Errorf("error writing response: %v", err)
		}
	}
}
