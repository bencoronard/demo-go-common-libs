package otel

import (
	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel/trace"
)

func RouteTagMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		req := c.Request()
		span := trace.SpanFromContext(req.Context())

		if span != nil {
			route := c.Path()
			if route != "" {
				span.SetName(req.Method + " " + route)
			}
		}

		return next(c)
	}
}
