package core

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// var SocketServer *socketio.Server

func MakeRoutes(e *echo.Echo) {
	// socket.io server connection
	// SocketServer = MakeSocketServer()

	// stripe webhook listener
	e.Any("/webhook", echo.HandlerFunc(HandleStripeWebhook))

	// token not required group
	e.POST("/login", Login)
	e.POST("/logout", Logout)
	// for donation
	// e.POST("/stripe/checkoutSession", CreateCheckoutSession)
	e.POST("/recover/:username", SendPasswordReset)
	e.POST("/recover/submit", ChangePassword)

	e.POST("/beta/request", SendBetaInviteRequest)

	// beta key required
	b := e.Group("")
	b.Use(HasBetaKey)
	b.POST("/user", CreateUser)

	// userSelector required
	u := e.Group("")
	u.Use(UserExists)
	u.Any("/ws/:userSelector", echo.HandlerFunc(WsEndpoint))

	mySigningKey := GetSigningKey()

	// r: requires token
	r := e.Group("/a")
	config := middleware.JWTConfig{
		Claims:     &jwtCustomClaims{},
		SigningKey: mySigningKey,
	}
	r.Use(middleware.JWTWithConfig(config))
	r.GET("/ping", Pong)

	r.GET("/user", GetUser)
	r.PATCH("/user", UpdateUser)
	r.PATCH("/avatar", UpdateAvatar)
	r.GET("/avatar", GetAvatar)
	r.GET("/pod/list", GetPodList)
	r.POST("/pod", CreatePod)
	r.POST("/pod/join", JoinPod)
	r.POST("/pod/join/summary", GetJoinPod)
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
	s.DELETE("/pod/leave/:podSelector/:selector", LeavePod)

	// v: require token, pod collaborator, and admin
	v := s.Group("")
	v.Use(IsAdmin)
	v.POST("/pod/invite/:podSelector", SendInvite)
	v.DELETE("/pod/invite/:podSelector/:selector", DeleteInvite)
	v.PATCH("/collaborator/role/:podSelector", UpdateCollaboratorRole)
	v.DELETE("/collaborator/:podSelector/:selector", DeleteCollaborator)
	v.PATCH("/pod/destroy/:podSelector", ScheduleDestroyPod)
	v.PATCH("/pod/:podSelector", UpdatePod)

	// a: require token, stripe account, pod collaborator
	a := s.Group("")
	a.Use(HasStripeAccount)
	a.POST("/stripe/checkoutsession/:podSelector", CreateCheckoutSession)
	a.POST("/stripe/refund/:podSelector/:txnId", ScheduleRefund)
	a.POST("/stripe/refund/cancel/:podSelector/:txnId", CancelScheduledRefund)

	// super: requires token and superadmin
	super := r.Group("")
	super.Use(IsSuperAdmin)
	super.POST("/beta/invite", SendBetaInvite)
	super.GET("/beta/requestlist", GetBetaRequestList)

}
