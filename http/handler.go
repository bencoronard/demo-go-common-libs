package http

import "github.com/labstack/echo/v5"

type ErrorHandlerFunc func(c *echo.Context, err error) error

func GlobalErrorHandler(fn ErrorHandlerFunc) func(c *echo.Context, err error) {
	return func(c *echo.Context, err error) {
		if resp, uErr := echo.UnwrapResponse(c.Response()); uErr == nil {
			if resp.Committed {
				return
			}
		}

		if fn != nil {
			if err = fn(c, err); err != nil {
				c.JSON(500, "Internal server error")
				return
			}
		}

		handleError(c, err)
	}
}

func handleError(c *echo.Context, err error) {
	c.JSON(500, "Internal server error")
}
