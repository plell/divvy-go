package userHandler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Oi(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func Outer(c echo.Context) error {
	return c.String(http.StatusOK, "nonono here!")
}
