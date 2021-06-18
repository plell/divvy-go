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

func IsApp(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(*jwtCustomClaims)
		isApp := claims.IsApp

		if !isApp {
			return c.String(http.StatusInternalServerError, "Action unauthorized.")
		}

		return next(c)
	}
}

func IsStore(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(*jwtCustomClaims)
		isStore := claims.IsStore

		if !isStore {
			return c.String(http.StatusInternalServerError, "Action unauthorized.")
		}

		return next(c)
	}
}

func IsSuperAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(*jwtCustomClaims)
		user_id := claims.UserID
		log.Println("IsSuperAdmin?")

		myuser := User{}
		result := DB.First(&myuser, user_id)
		if result.Error != nil {
			return AbstractError(c, "Can't find user")
		}

		if myuser.UserTypeID != USER_TYPE_SUPER {
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
			return c.String(http.StatusInternalServerError, "You haven't linked a deposit account.")
		}

		return next(c)
	}
}

// requires podSelector in route url
func IsPodMember(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Println("IsPodMember?")

		user := c.Get("user").(*jwt.Token)
		claims := user.Claims.(*jwtCustomClaims)
		user_id := claims.UserID

		podSelector := c.Param("podSelector")

		pod := Pod{}
		result := DB.Where("selector = ?", podSelector).First(&pod)
		if result.Error != nil {
			return AbstractError(c, "Can't find pod")
		}

		collaborator := Collaborator{}
		result = DB.Where("user_id = ?", user_id).Where("pod_id = ?", pod.ID).First(&collaborator)
		if result.Error != nil {
			return c.String(http.StatusInternalServerError, "You're not a member of this wallet")
		}

		return next(c)
	}
}

// requires posSelector in route url
func PodIsNotScheduledForDelete(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		podSelector := c.Param("podSelector")
		pod := Pod{}
		result := DB.Where("selector = ?", podSelector).First(&pod)
		if result.Error != nil {
			return c.String(http.StatusInternalServerError, "Cannot find wallet")
		}

		if pod.ToDelete != "" {
			return c.String(http.StatusInternalServerError, "This wallet is scheduled for delete.")
		}

		return next(c)
	}
}

func UserExists(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userSelector := c.Param("userSelector")
		log.Println("UserExists?")
		user := User{}
		result := DB.Where("selector = ?", userSelector).First(&user)

		if result.Error != nil {
			return c.String(http.StatusUnauthorized, "Unauthorized")
		}

		return next(c)
	}
}
