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
	Handle(err error, pd dto.ProblemDetail) (dto.ProblemDetail, bool)
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

		pd := dto.NewProblemDetail(status).
			WithDetail(detail).
			With("timestamp", time.Now()).
			With("trace", h.extractTraceID(c.Request().Context()))

		if h.ah != nil {
			pd, handled = h.ah.Handle(err, pd)
		}
		if !handled {
			pd = handleUnhandledError(err, pd)
		}

		c.JSON(pd.Status(), pd)
	}
}

func handleUnhandledError(err error, pd dto.ProblemDetail) dto.ProblemDetail {
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
		return pd.
			WithStatus(http.StatusBadRequest).
			WithDetail("Input data did not pass validations").
			With("errors", validationDetails)
	case errors.Is(err, auth.ErrInsufficientPermission):
		return pd.
			WithStatus(http.StatusForbidden).
			WithDetail(err.Error())
	case errors.Is(err, ErrMissingRequestHeader):
		return pd.
			WithStatus(http.StatusBadRequest).
			WithDetail(err.Error())
	case errors.Is(err, jwt.ErrTokenVerificationFail),
		errors.Is(err, jwt.ErrTokenClaimsInvalid),
		errors.Is(err, jwt.ErrTokenMalformed):
		return pd.
			WithStatus(http.StatusUnauthorized).
			WithDetail(err.Error())
	default:
		slog.Error("unhandled error", "error", err)
		return pd
	}
}
