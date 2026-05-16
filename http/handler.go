package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/bencoronard/demo-go-common-libs/auth"
	"github.com/bencoronard/demo-go-common-libs/dto"
	"github.com/bencoronard/demo-go-common-libs/validator"
	"github.com/labstack/echo/v5"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type globalErrorHandler struct {
	appErrHandler  AppErrorHandler
	tracerProvider *sdktrace.TracerProvider
}

func (h *globalErrorHandler) GetHandler() func(c *echo.Context, err error) {
	return func(c *echo.Context, err error) {
		if resp, err := echo.UnwrapResponse(c.Response()); err == nil && resp.Committed {
			return
		}

		status := http.StatusInternalServerError
		detail := "Unhandled error at server side"
		handled := false

		var sc echo.HTTPStatusCoder
		if errors.As(err, &sc) && sc.StatusCode() != 0 {
			status = sc.StatusCode()
			detail = err.Error()
			handled = true
		}

		pd := dto.NewProblemDetail(status).
			WithDetail(detail).
			With("timestamp", time.Now())

		if h.tracerProvider != nil {
			pd = pd.With("trace", extractTraceID(c.Request().Context()))
		}

		if h.appErrHandler != nil {
			pd, handled = h.appErrHandler.Handle(err, pd)
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
	ve, ok := errors.AsType[*validator.ValidationError](err)
	if ok {
		return pd.
			WithStatus(http.StatusBadRequest).
			WithDetail(ve.Error()).
			With("errors", ve.Errors)
	}

	switch {
	case errors.Is(err, auth.ErrOperationNotPermitted):
		return pd.
			WithStatus(http.StatusForbidden).
			WithDetail(err.Error())
	case errors.Is(err, ErrAuthHeaderInvalid),
		errors.Is(err, ErrAuthTokenInvalid):
		return pd.
			WithStatus(http.StatusUnauthorized).
			WithDetail(err.Error())
	default:
		return pd
	}
}

func extractTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx).SpanContext()
	if !span.IsValid() {
		return ""
	}
	return span.TraceID().String()
}
