package core

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func MakeRoutes(e *echo.Echo) {
	// token not required group
	e.POST("/login", Login)
	e.POST("/user", CreateUser)
	e.POST("/stripe/checkoutSession", CreateCheckoutSession)
	e.POST("/passwordr/:username", SendPasswordReset)

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
	r.GET("/pod/list", GetPodList)
	r.GET("/pod/:selector", GetPod)
	r.GET("/pod/invitelist/:podSelector", GetInvites)
	r.POST("/pod", CreatePod)
	r.POST("/pod/join", JoinPod)
	r.POST("/pod/invite", SendInvite)
	r.DELETE("/pod/invite/:selector", DeleteInvite)

	r.GET("/collaboratorlist/:podSelector", GetCollaboratorList)
	r.PATCH("/collaborator/admin", UpdateCollaboratorAdmin)
	r.DELETE("/collaborator/:selector", DeleteCollaborator)

	r.POST("/stripe/account", LinkStripeAccount)
	r.POST("/stripe/checkoutsession", CreateCheckoutSession)
	r.POST("/stripe/refund/:txnId", ScheduleRefund)
	r.POST("/stripe/refund/cancel/:txnId", CancelScheduledRefund)
	r.GET("/stripe/transferlist/:podSelector", GetPodTransferList)
	// we may not be able to get payouts for individual accounts...
	r.GET("/stripe/payoutlist/:podSelector", GetPodPayoutList)
	r.GET("/stripe/chargelist/:podSelector", GetPodChargeList)
	r.GET("/stripe/account", GetStripeAccount)

	// e.POST("/users", createUser)
	// e.GET("/users/:id", getUser)
	// e.PUT("/users/:id", updateUser)
	// e.DELETE("/users/:id", deleteUser)
}
