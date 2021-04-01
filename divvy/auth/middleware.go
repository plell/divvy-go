package auth

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// var mySigningKey = os.Get("MY_JWT_TOKEN") // get this from env
var mySigningKey = []byte("mysecretphrase")

var IsLoggedIn = middleware.JWTWithConfig(middleware.JWTConfig{
	SigningKey: mySigningKey})

func IsAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		isAdmin := claims["admin"].(bool)
		if isAdmin == false {
			return echo.ErrUnauthorized
		}
		return next(c)
	}
}
