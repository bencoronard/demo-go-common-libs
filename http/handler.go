package http

import (
	"context"
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
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

type AppErrorHandler interface {
	Handle(err error, pd *dto.ProblemDetail) error
}

type GlobalErrorHandlerParams struct {
	fx.In
	AppErrHandler  AppErrorHandler          `optional:"true"`
	TracerProvider *sdktrace.TracerProvider `optional:"true"`
}

type GlobalErrorHandler interface {
	GetHandler() func(c *echo.Context, err error)
}

type globalErrorHandlerImpl struct {
	ah AppErrorHandler
	tp *sdktrace.TracerProvider
}

func NewGlobalErrorHandler(p GlobalErrorHandlerParams) GlobalErrorHandler {
	return &globalErrorHandlerImpl{
		ah: p.AppErrHandler,
		tp: p.TracerProvider,
	}
}

func (h *globalErrorHandlerImpl) extractTraceID(ctx context.Context) string {
	if h.tp == nil {
		return ""
	}
	span := trace.SpanFromContext(ctx).SpanContext()
	if span.IsValid() {
		return span.TraceID().String()
	}
	return ""
}

func (h *globalErrorHandlerImpl) GetHandler() func(c *echo.Context, err error) {
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
				"trace":     h.extractTraceID(c.Request().Context()),
			},
		}

		if h.ah != nil {
			handled = h.ah.Handle(err, &pd) == nil
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
		pd.Detail = "Input data did not pass validations"
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
