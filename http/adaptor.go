package http

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type AuthenticatedHandler func(
	http.ResponseWriter,
	*http.Request,
	jwt.MapClaims,
)

type Adapter struct {
	resolver HttpAuthHeaderResolver
}

func NewAdapter(resolver HttpAuthHeaderResolver) *Adapter {
	return &Adapter{resolver: resolver}
}

func (a *Adapter) Wrap(h AuthenticatedHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		claims, err := a.resolver.ExtractClaims(r)
		if err != nil {
			writeAuthError(w, err)
			return
		}

		h(w, r, claims)
	})
}

func writeAuthError(w http.ResponseWriter, err error) {
	switch err {
	case ErrMissingRequestHeader:
		http.Error(w, err.Error(), http.StatusUnauthorized)
	case jwt.ErrTokenMalformed:
		http.Error(w, err.Error(), http.StatusUnauthorized)
	default:
		http.Error(w, "authentication failed", http.StatusUnauthorized)
	}
}
