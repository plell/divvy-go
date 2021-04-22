package core

import (
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
)

func HasStripeAccount(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(*jwtCustomClaims)
		user_id := claims.UserID
		log.Println("HasStripeAccount?")
		stripeAccount := StripeAccount{}
		result := DB.Where("user_id = ?", user_id).First(&stripeAccount)

		if result.Error != nil {
			return c.String(http.StatusInternalServerError, "Finish account setup first")
		}

		return next(c)
	}
}

// requires posSelector in route url
func IsPodMember(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(*jwtCustomClaims)
		user_id := claims.UserID
		log.Println("IsPodMember?")
		podSelector := c.Param("podSelector")
		pod := Pod{}
		result := DB.Where("selector = ?", podSelector).First(&pod)
		if result.Error != nil {
			return c.String(http.StatusInternalServerError, "Cannot find wallet, reload page")
		}

		collaborator := Collaborator{}
		result = DB.Where("user_id = ?", user_id).Where("pod_id = ?", pod.ID).First(&collaborator)
		if result.Error != nil {
			return c.String(http.StatusInternalServerError, "You're not a member of this wallet")
		}

		return next(c)
	}
}
