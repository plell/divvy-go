package routes

import (
	"github.com/labstack/echo/v4"
	// "github.com/plell/divvygo/divvy/handlers/userHandler"
	user "github.com/plell/divvygo/divvy/handlers/userHandler"
)

func MakeRoutes(e *echo.Echo) {
	// Routes
	e.GET("/", user.Okokok)
	e.GET("/user", user.Hello)

	// e.POST("/users", createUser)
	// e.GET("/users/:id", getUser)
	// e.PUT("/users/:id", updateUser)
	// e.DELETE("/users/:id", deleteUser)
}
