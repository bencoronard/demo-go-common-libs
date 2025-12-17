package http

import (
	"encoding/json"
	"net/http"

	"github.com/bencoronard/demo-go-common-libs/rfc9457"
)

type ResponseErrorHandler interface {
	RespondWithError(w http.ResponseWriter, err error) bool
}

type responseErrorHandlerImpl struct {
	h ResponseErrorHandler
}

func NewResponseErrorHandlerImpl() ResponseErrorHandler {
	return &responseErrorHandlerImpl{}
}

func (r *responseErrorHandlerImpl) RespondWithError(w http.ResponseWriter, err error) bool {
	if ok := r.h.RespondWithError(w, err); ok {
		return true
	}
	writeError(w, err)
	return true
}

func writeError(w http.ResponseWriter, err error) {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	w.Header().Set("Content-Type", "application/problem+json")

	switch err {
	case ErrMissingRequestHeader:
		w.WriteHeader(http.StatusUnauthorized)
		enc.Encode(rfc9457.ForStatusAndDetail(http.StatusUnauthorized, "Missing required request headers"))
	default:
		w.WriteHeader(http.StatusInternalServerError)
		enc.Encode(rfc9457.ForStatusAndDetail(http.StatusInternalServerError, "Unhandled exception at server side"))
	}
}
