package userHandler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func Okokok(c echo.Context) error {
	return c.String(http.StatusOK, "nonono here!")
}
