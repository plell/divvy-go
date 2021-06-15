package core

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// var SocketServer *socketio.Server

func MakeRoutes(e *echo.Echo) {

	e.GET("/", func(c echo.Context) error {
		return c.HTML(http.StatusOK, `
			<h1>Welcome to Echo!</h1>
			<h3>TLS certificates automatically installed from Let's Encrypt :)</h3>
		`)
	})

	// stripe webhook listener
	e.Any("/webhook", echo.HandlerFunc(HandleStripeWebhook))

	// token not required group
	e.GET("/avatarOptions", GetAvatarOptions)
	e.POST("/customerLogin", CustomerLogin)
	e.POST("/login", Login)
	e.POST("/logout", Logout)
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

	r.POST("/pod/join/summary", GetJoinPod)
	r.POST("/stripe/account", LinkStripeAccount)
	r.GET("/stripe/account", GetStripeAccount)
	r.POST("/verify/:verificationCode", VerifyAccountEmail)
	r.POST("/sendVerification", SendVerificationEmail)
	r.POST("/user/transfers/:userSelector", GetUserTransfers)
	r.POST("/user/transferlist/:userSelector", GetUserTransferList)

	r.POST("/pod/join", JoinPod)

	r.GET("/stripe/customerAccount", GetStripeCustomerAccount)
	r.GET("/stripe/customerPortal", CreateCustomerPortalSession)
	r.PATCH("/stripe/customerCheckoutSession/:sessionId", UpdateCheckoutSessionByCustomer)

	// s: require token, pod collaborator
	s := r.Group("")
	s.Use(IsPodMember)
	s.GET("/pod/:podSelector", GetPod)
	s.GET("/pod/invitelist/:podSelector", GetInvites)
	s.GET("/pod/unavailableCharges/:podSelector", GetPodUnavailableChargeList)
	s.POST("/pod/transfers/:podSelector", GetPodTransfers)
	s.DELETE("/pod/leave/:podSelector/:selector", LeavePod)
	s.GET("/stripe/transferlist/:podSelector", GetPodTransferList)
	s.GET("/stripe/payoutlist/:podSelector", GetPodPayoutList)
	s.POST("/stripe/chargelist/:podSelector", GetPodChargeList)
	s.GET("/collaboratorlist/:podSelector", GetCollaboratorList)

	// v: require token, pod collaborator, and admin
	v := s.Group("")
	v.Use(IsAdmin)
	v.PATCH("/pod/restore/:podSelector", CancelScheduleDestroyPod)
	v.DELETE("/pod/invite/:podSelector/:selector", DeleteInvite)

	// vp: require token, pod collaborator, admin, pod cant be schedule for delete
	vp := v.Group("")
	vp.Use(PodIsNotScheduledForDelete)
	vp.PATCH("/pod/destroy/:podSelector", ScheduleDestroyPod)
	vp.PATCH("/pod/:podSelector", UpdatePod)
	vp.POST("/pod/invite/:podSelector", SendInvite)
	vp.PATCH("/collaborator/role/:podSelector", UpdateCollaboratorRole)
	vp.DELETE("/collaborator/:podSelector/:selector", DeleteCollaborator)

	// a: require token, stripe account, pod collaborator
	a := s.Group("")
	a.Use(HasStripeAccount)
	a.POST("/stripe/refund/:podSelector/:txnId", ScheduleRefund)
	a.POST("/stripe/refund/cancel/:podSelector/:txnId", CancelScheduledRefund)

	// a: require token, stripe account, pod collaborator, pod not deleting
	ap := a.Group("")
	ap.Use(PodIsNotScheduledForDelete)
	ap.POST("/stripe/checkoutSession/:podSelector", CreateCheckoutSession)

	// super: requires token and superadmin
	super := r.Group("")
	super.Use(IsSuperAdmin)
	super.POST("/beta/invite", SendBetaInvite)
	super.GET("/beta/requestlist", GetBetaRequestList)

}
