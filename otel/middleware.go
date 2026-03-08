package otel

import (
	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
)

func RouteTagMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		err := next(c)
		labeler, _ := otelhttp.LabelerFromContext(c.Request().Context())
		labeler.Add(attribute.String("http.route", c.Path()))
		return err
	}
}
