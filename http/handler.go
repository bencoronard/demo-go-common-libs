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
	"github.com/bencoronard/demo-go-common-libs/validator"
	"github.com/labstack/echo/v5"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type globalErrorHandler struct {
	ah AppErrorHandler
	tp *sdktrace.TracerProvider
}

func (h *globalErrorHandler) GetHandler() func(c *echo.Context, err error) {
	return func(c *echo.Context, err error) {
		if resp, err := echo.UnwrapResponse(c.Response()); err == nil && resp.Committed {
			return
		}

		status := http.StatusInternalServerError
		detail := "Unhandled error at server side"
		handled := false

		sc, ok := errors.AsType[*echo.HTTPError](err)
		if ok && sc.StatusCode() != 0 {
			status = sc.StatusCode()
			detail = err.Error()
			handled = true
		}

		pd := dto.NewProblemDetail(status).
			WithDetail(detail).
			With("timestamp", time.Now()).
			With("trace", h.extractTraceId(c.Request().Context()))

		if h.ah != nil {
			pd, handled = h.ah.Handle(err, pd)
		}
		if !handled {
			pd = handleUnhandledError(err, pd)
		}

		if pd.Title() == "" {
			pd = pd.WithTitle(http.StatusText(pd.Status()))
		}

		c.JSON(pd.Status(), pd)
	}
}

func handleUnhandledError(err error, pd dto.ProblemDetail) dto.ProblemDetail {
	var ve validator.ValidationError
	switch {
	case errors.As(err, &ve):
		return pd.
			WithStatus(http.StatusBadRequest).
			WithDetail(ve.Error()).
			With("errors", ve.Data())
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

func (h *globalErrorHandler) extractTraceId(ctx context.Context) string {
	if h.tp == nil {
		return ""
	}
	span := trace.SpanFromContext(ctx).SpanContext()
	if span.IsValid() {
		return span.TraceID().String()
	}
	return ""
}
