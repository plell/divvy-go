package routes

import (
	"github.com/labstack/echo/v4"

	auth "github.com/plell/divvygo/divvy/auth"
	db "github.com/plell/divvygo/divvy/database"
)

func MakeRoutes(e *echo.Echo) {
	// auth routes
	e.POST("/login", auth.Login)

	// user routes
	e.POST("/createUser", db.CreateUser, auth.IsLoggedIn)
	e.GET("/user/:userId", db.GetUser, auth.IsLoggedIn)

	// e.POST("/users", createUser)
	// e.GET("/users/:id", getUser)
	// e.PUT("/users/:id", updateUser)
	// e.DELETE("/users/:id", deleteUser)
}
