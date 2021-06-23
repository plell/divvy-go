package core

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// var SocketServer *socketio.Server

func MakeRoutes(e *echo.Echo) {

	e.Use(LogPathAndIp)
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

	e.POST("/login", Login)
	e.POST("/logout", Logout)
	e.POST("/recover/:username", SendPasswordReset)
	e.POST("/recover/submit", ChangePassword)
	e.POST("/user", CreateUser)
	e.POST("/googleLogin", GoogleLoginOrSignUp)
	e.POST("/beta/request", SendBetaInviteRequest)

	e.POST("/customerLogin", CustomerLogin)
	e.POST("/customerUser", CustomerCreateUser)
	e.POST("/customerGoogleLogin", CustomerGoogleLoginOrSignUp)

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

	storeUser := r.Group("")
	storeUser.Use(IsStore)
	storeUser.GET("/stripe/customerAccount", GetStripeCustomerAccount)
	storeUser.GET("/stripe/customerPortal", CreateCustomerPortalSession)
	storeUser.PATCH("/stripe/customerCheckoutSession/:sessionId", UpdateCheckoutSessionByCustomer)

	appUser := r.Group("")
	appUser.Use(IsApp)
	appUser.PATCH("/avatar", UpdateAvatar)
	appUser.GET("/avatar", GetAvatar)
	appUser.GET("/pod/list", GetPodList)
	appUser.POST("/pod", CreatePod)

	appUser.POST("/pod/join/summary", GetJoinPod)
	appUser.POST("/stripe/account", LinkStripeAccount)
	appUser.GET("/stripe/account", GetStripeAccount)
	appUser.POST("/verify/:verificationCode", VerifyAccountEmail)
	appUser.POST("/sendVerification", SendVerificationEmail)
	appUser.POST("/user/transfers/:userSelector", GetUserTransfers)
	appUser.POST("/user/transferlist/:userSelector", GetUserTransferList)
	appUser.POST("/pod/join", JoinPod)
	appUser.PATCH("/user", UpdateUser)

	// s: require token, appUser, pod collaborator
	s := appUser.Group("")
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

	// v: require token, appUser, pod collaborator, and admin
	v := s.Group("")
	v.Use(IsAdmin)
	v.PATCH("/pod/restore/:podSelector", CancelScheduleDestroyPod)
	v.DELETE("/pod/invite/:podSelector/:selector", DeleteInvite)

	// vp: require token, appUser, pod collaborator, admin, pod cant be schedule for delete
	vp := v.Group("")
	vp.Use(PodIsNotScheduledForDelete)
	vp.PATCH("/pod/destroy/:podSelector", ScheduleDestroyPod)
	vp.PATCH("/pod/:podSelector", UpdatePod)
	vp.POST("/pod/invite/:podSelector", SendInvite)
	vp.PATCH("/collaborator/role/:podSelector", UpdateCollaboratorRole)
	vp.DELETE("/collaborator/:podSelector/:selector", DeleteCollaborator)

	// a: require token, appUser, stripe account, pod collaborator
	a := s.Group("")
	a.Use(HasStripeAccount)
	a.POST("/stripe/refund/:podSelector/:txnId", ScheduleRefund)
	a.POST("/stripe/refund/cancel/:podSelector/:txnId", CancelScheduledRefund)

	// a: require token, appUser, stripe account, pod collaborator, pod not deleting
	ap := a.Group("")
	ap.Use(PodIsNotScheduledForDelete)
	ap.POST("/stripe/checkoutSession/:podSelector", CreateCheckoutSession)

	// super: requires token, appUser, and superadmin
	super := appUser.Group("")
	super.Use(IsSuperAdmin)
	super.POST("/beta/invite", SendBetaInvite)
	super.GET("/beta/requestlist", GetBetaRequestList)

}
