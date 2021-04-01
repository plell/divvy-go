package database

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func abstractError(c echo.Context) error {
	return c.String(http.StatusInternalServerError, "")
}
