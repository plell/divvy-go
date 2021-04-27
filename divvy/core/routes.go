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
	e.POST("/recover/:username", SendPasswordReset)
	e.POST("/recover/submit", ChangePassword)

	mySigningKey := GetSigningKey()

	// r: requires token
	r := e.Group("/a")
	config := middleware.JWTConfig{
		Claims:     &jwtCustomClaims{},
		SigningKey: mySigningKey,
	}
	r.Use(middleware.JWTWithConfig(config))
	r.GET("/ping", Pong)
	e.POST("/logout", Logout)
	r.GET("/user", GetUser)
	r.PATCH("/avatar", UpdateAvatar)
	r.GET("/avatar", GetAvatar)
	r.GET("/pod/list", GetPodList)
	r.POST("/pod", CreatePod)
	r.POST("/pod/join", JoinPod)
	r.POST("/stripe/account", LinkStripeAccount)
	r.GET("/stripe/account", GetStripeAccount)
	r.POST("/verify/:verificationCode", VerifyAccountEmail)
	r.POST("/sendVerification", SendVerificationEmail)

	// s: require token, pod collaborator
	s := r.Group("")
	s.Use(IsPodMember)
	s.GET("/pod/:podSelector", GetPod)
	s.GET("/pod/invitelist/:podSelector", GetInvites)
	s.GET("/stripe/transferlist/:podSelector", GetPodTransferList)
	s.GET("/stripe/payoutlist/:podSelector", GetPodPayoutList)
	s.GET("/stripe/chargelist/:podSelector", GetPodChargeList)
	s.GET("/collaboratorlist/:podSelector", GetCollaboratorList)
	s.POST("/pod/invite/:podSelector", SendInvite)
	s.DELETE("/pod/invite/:podSelector/:selector", DeleteInvite)

	s.PATCH("/collaborator/admin/:podSelector", UpdateCollaboratorAdmin)
	s.DELETE("/collaborator/:podSelector/:selector", DeleteCollaborator)

	// a: require token, stripe account, pod collaborator
	a := s.Group("")
	a.Use(HasStripeAccount)
	a.POST("/stripe/checkoutsession/:podSelector", CreateCheckoutSession)
	a.POST("/stripe/refund/:podSelector/:txnId", ScheduleRefund)
	a.POST("/stripe/refund/cancel/:podSelector/:txnId", CancelScheduledRefund)

}
