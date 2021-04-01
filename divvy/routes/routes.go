package routes

import (
	"github.com/labstack/echo/v4"
	auth "github.com/plell/divvygo/divvy/auth"
	db "github.com/plell/divvygo/divvy/database"
	misc "github.com/plell/divvygo/divvy/misc"
)

func MakeRoutes(e *echo.Echo) {
	// tokenless routes
	e.POST("/login", auth.Login)
	e.POST("/createUser", db.CreateUser)

	// token required (auth.IsLoggedIn middleware)
	e.GET("/ping", misc.Pong, auth.IsLoggedIn)
	e.GET("/user/:userId", db.GetUser, auth.IsLoggedIn)

	// e.POST("/users", createUser)
	// e.GET("/users/:id", getUser)
	// e.PUT("/users/:id", updateUser)
	// e.DELETE("/users/:id", deleteUser)
}
