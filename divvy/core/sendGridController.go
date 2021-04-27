// using SendGrid's Go Library
// https://github.com/sendgrid/sendgrid-go
package core

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var SENDGRID_INVITE_TEMPLATE = "d-4416cc09847445c9867d1e9d3cf09dcc"
var SENDGRID_VERIFICATION_TEMPLATE = "d-c0ae63959f8c4a30aca48f6599b07ed4"
var SENDGRID_REFUND_LIMIT_TEMPLATE = "d-4416cc09847445c9867d1e9d3cf09dcc"
var SENDGRID_PW_RESET_TEMPLATE = "d-f05cdb7f762e48d0a1188f3b0173163d"

type InviteCreator struct {
	Email       string `json:"email"`
	PodSelector string `json:"podSelector"`
}

func SendInvite(c echo.Context) error {
	user_id, err := GetUserIdFromToken(c)
	if err != nil {
		return AbstractError(c, "Something went wrong")
	}

	req := InviteCreator{}
	defer c.Request().Body.Close()
	err = json.NewDecoder(c.Request().Body).Decode(&req)

	if err != nil {
		return AbstractError(c, "Something went wrong")
	}

	// get pod
	pod := Pod{}
	result := DB.Where("selector = ?", req.PodSelector).First(&pod)

	if result.Error != nil {
		return AbstractError(c, "Something went wrong")
	}

	code := MakeInviteCode()

	invite := Invite{
		Code:        code,
		Email:       req.Email,
		PodID:       pod.ID,
		CreatedByID: user_id,
		Selector:    MakeSelector(INVITE_TABLE),
	}

	result = DB.Create(&invite)

	if result.Error != nil {
		return AbstractError(c, "Something went wrong")
	}

	user := User{}

	result = DB.First(&user, user_id)

	if result.Error != nil {
		return AbstractError(c, "Something went wrong")
	}

	SendInviteEmail(user.DisplayName, req.Email, code)

	return c.String(http.StatusOK, "Success!")
}

func SendInviteEmail(senderName string, email string, inviteCode string) {
	m := mail.NewV3Mail()

	address := "invited@jamwallet.com"
	name := "Jamwallet"
	e := mail.NewEmail(name, address)
	m.SetFrom(e)

	m.SetTemplateID(SENDGRID_INVITE_TEMPLATE)

	p := mail.NewPersonalization()
	tos := []*mail.Email{
		mail.NewEmail("", email),
	}

	p.AddTos(tos...)

	p.SetDynamicTemplateData("inviteCode", inviteCode)
	p.SetDynamicTemplateData("senderName", senderName)

	m.AddPersonalizations(p)

	request := sendgrid.GetRequest(os.Getenv("SENDGRID_API_KEY"), "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	var Body = mail.GetRequestBody(m)
	request.Body = Body
	response, err := sendgrid.API(request)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
}

func SendPasswordReset(c echo.Context) error {
	username := c.Param("username")

	user := User{}

	result := DB.Where("username = ?", username).First(&user)
	if result.Error != nil {
		return AbstractError(c, "")
	}
	code := MakeInviteCode()
	user.PasswordResetToken = code
	result = DB.Save(&user)
	if result.Error != nil {
		return AbstractError(c, "Couldn't save")
	}

	SendPasswordResetEmail(username, code)
	// always return success, to avoid letting people fish for accounts
	return c.String(http.StatusOK, "Success!")
}

func SendPasswordResetEmail(username string, code string) {
	m := mail.NewV3Mail()

	address := "request@jamwallet.com"
	name := "Jamwallet"
	e := mail.NewEmail(name, address)
	m.SetFrom(e)

	m.SetTemplateID(SENDGRID_PW_RESET_TEMPLATE)

	p := mail.NewPersonalization()
	tos := []*mail.Email{
		mail.NewEmail("", username),
	}

	p.AddTos(tos...)

	p.SetDynamicTemplateData("resetCode", code)

	m.AddPersonalizations(p)

	request := sendgrid.GetRequest(os.Getenv("SENDGRID_API_KEY"), "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	var Body = mail.GetRequestBody(m)
	request.Body = Body
	response, err := sendgrid.API(request)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
}

func SendRefundLimitEmail(pod Pod) {

}

func FwdToSendVerificationEmail(c echo.Context) error {
	user_id, err := GetUserIdFromToken(c)
	if err != nil {
		return AbstractError(c, "Something went wrong")
	}
	user := User{}
	result := DB.Find(&user, user_id)
	if result.Error != nil {
		return AbstractError(c, "Couldn't find user")
	}

	return c.String(http.StatusOK, "Sent verification email to "+user.Username)
}

func SendVerificationEmail(user User) {

	log.Println("SendVerificationEmail")

	m := mail.NewV3Mail()

	address := "verification@jamwallet.com"
	name := "jamWallet"
	e := mail.NewEmail(name, address)
	m.SetFrom(e)

	m.SetTemplateID(SENDGRID_VERIFICATION_TEMPLATE)

	p := mail.NewPersonalization()
	tos := []*mail.Email{
		mail.NewEmail("", user.Username),
	}

	p.AddTos(tos...)

	verificationCode := EmailVerificationCode{}

	result := DB.Where("user_id = ?", user.ID).First(&verificationCode)
	if result.Error != nil {
		// make verification code
		verificationCode = EmailVerificationCode{
			UserID: user.ID,
			Code:   MakeInviteCode(),
		}
		DB.Create(&verificationCode)
	}

	p.SetDynamicTemplateData("verificationCode", verificationCode.Code)

	m.AddPersonalizations(p)

	request := sendgrid.GetRequest(os.Getenv("SENDGRID_API_KEY"), "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	var Body = mail.GetRequestBody(m)
	request.Body = Body
	response, err := sendgrid.API(request)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
}
