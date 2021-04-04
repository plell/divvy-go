package core

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func MakeRoutes(e *echo.Echo) {
	// token not required group
	e.POST("/login", Login)
	e.POST("/user", CreateUser)

	mySigningKey := GetSigningKey()

	// token required group
	r := e.Group("/a")
	config := middleware.JWTConfig{
		Claims:     &jwtCustomClaims{},
		SigningKey: mySigningKey,
	}
	r.Use(middleware.JWTWithConfig(config))

	r.GET("/ping", Pong)
	r.GET("/user/:userId", GetUser)
	r.PATCH("/avatar", UpdateAvatar)
	r.GET("/avatar", GetAvatar)
	r.GET("/pod/list", GetDivvyPodList)
	r.GET("/pod/:selector", GetDivvyPod)
	r.POST("/pod", CreateDivvyPod)
	r.POST("/stripe/account", CreateStripeAccount)
	r.GET("/stripe/account", GetStripeAccount)
	// e.POST("/users", createUser)
	// e.GET("/users/:id", getUser)
	// e.PUT("/users/:id", updateUser)
	// e.DELETE("/users/:id", deleteUser)
}
