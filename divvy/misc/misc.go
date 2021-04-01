package misc

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Pong(c echo.Context) error {
	return c.String(http.StatusOK, "Pong")
}
