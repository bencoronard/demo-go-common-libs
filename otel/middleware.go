package otel

import (
	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func RouteTagMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		err := next(c)

		span := trace.SpanFromContext(c.Request().Context())
		if span.IsRecording() {
			span.SetName(c.Request().Method + " " + c.Path())
			span.SetAttributes(attribute.String("http.route", c.Path()))
		}

		return err
	}
}
