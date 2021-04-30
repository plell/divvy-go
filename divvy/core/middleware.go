package core

import (
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
)

func HasBetaKey(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// req := UserCreator{}
		// defer c.Request().Body.Close()
		// err := json.NewDecoder(c.Request().Body).Decode(&req)

		// if err != nil {
		// 	return AbstractError(c, "Couldn't read request")
		// }

		// if req.BetaKey == "" {
		// 	return AbstractError(c, "You must have a beta key")
		// }

		// betaKey := BetaKey{}
		// result := DB.Where("beta_key = ?", req.BetaKey).First(&betaKey)
		// if result.Error != nil {
		// 	return AbstractError(c, "Beta key invalid")
		// }

		return next(c)
	}
}

func IsSuperAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(*jwtCustomClaims)
		selector := claims.UserSelector
		log.Println("IsSuperAdmin?")

		if selector != SUPERADMIN_SELECTOR {
			return c.String(http.StatusInternalServerError, "Action unauthorized.")
		}

		return next(c)
	}
}

func IsAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(*jwtCustomClaims)
		user_id := claims.UserID
		podSelector := c.Param("podSelector")
		log.Println("IsAdmin?")

		pod := Pod{}
		result := DB.Where("selector = ?", podSelector).First(&pod)
		if result.Error != nil {
			return AbstractError(c, "Can't find pod")
		}

		collaborator := Collaborator{}
		result = DB.Where("pod_id = ?", pod.ID).Where("user_id = ?", user_id).First(&collaborator)

		if collaborator.RoleTypeID != ROLE_TYPE_ADMIN {
			return c.String(http.StatusInternalServerError, "Action unauthorized.")
		}

		return next(c)
	}
}

func HasStripeAccount(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(*jwtCustomClaims)
		user_id := claims.UserID
		log.Println("HasStripeAccount?")
		stripeAccount := StripeAccount{}
		result := DB.Where("user_id = ?", user_id).First(&stripeAccount)

		if result.Error != nil {
			return c.String(http.StatusInternalServerError, "Your payouts are disabled. Create a Stripe account.")
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
