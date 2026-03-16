package http

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/bencoronard/demo-go-common-libs/auth"
	"github.com/bencoronard/demo-go-common-libs/dto"
	"github.com/bencoronard/demo-go-common-libs/jwt"
	"github.com/bencoronard/demo-go-common-libs/validation"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

type AppErrorHandlerFunc func(err error, pd *dto.ProblemDetail) error

func GlobalErrorHandler(fn AppErrorHandlerFunc) echo.HTTPErrorHandler {
	return func(c *echo.Context, err error) {
		if resp, err := echo.UnwrapResponse(c.Response()); err == nil && resp.Committed {
			return
		}

		status := http.StatusInternalServerError
		detail := "Unhandled error at server side"
		handled := false

		var sc echo.HTTPStatusCoder
		if errors.As(err, &sc) {
			if tmp := sc.StatusCode(); tmp != 0 {
				status = tmp
				detail = err.Error()
				handled = true
			}
		}

		pd := dto.ProblemDetail{
			Type:   "about:blank",
			Status: status,
			Detail: detail,
			Properties: map[string]any{
				"timestamp": time.Now(),
			},
		}

		if fn != nil {
			handled = fn(err, &pd) == nil
		}
		if !handled {
			handleUnhandledError(err, &pd)
		}

		if pd.Title == "" {
			pd.Title = http.StatusText(pd.Status)
		}

		c.JSON(pd.Status, pd)
	}
}

func handleUnhandledError(err error, pd *dto.ProblemDetail) {
	var ve validator.ValidationErrors
	switch {
	case errors.As(err, &ve):
		var validationDetails []validation.FieldValidationError
		for _, fe := range ve {
			validationDetails = append(validationDetails, validation.FieldValidationError{
				Field:   fe.Field(),
				Message: fe.Error(),
			})
		}
		if pd.Properties == nil {
			pd.Properties = make(map[string]any)
		}
		pd.Status = http.StatusBadRequest
		pd.Detail = err.Error()
		pd.Properties["errors"] = validationDetails
	case errors.Is(err, auth.ErrInsufficientPermission):
		pd.Status = http.StatusForbidden
		pd.Detail = err.Error()
	case errors.Is(err, ErrMissingRequestHeader):
		pd.Status = http.StatusBadRequest
		pd.Detail = err.Error()
	case errors.Is(err, jwt.ErrTokenVerificationFail),
		errors.Is(err, jwt.ErrTokenClaimsInvalid),
		errors.Is(err, jwt.ErrTokenMalformed):
		pd.Status = http.StatusUnauthorized
		pd.Detail = err.Error()
	default:
		slog.Error("unhandled error", "error", err)
	}
}
