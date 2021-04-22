// using SendGrid's Go Library
// https://github.com/sendgrid/sendgrid-go
package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var SENDGRID_INVITE_TEMPLATE = "d-4416cc09847445c9867d1e9d3cf09dcc"
var SENDGRID_PW_RESET_TEMPLATE = "d-4416cc09847445c9867d1e9d3cf09dcc"

func SendPasswordReset(c echo.Context) error {
	// username := c.Param("username")

	// user := User{}

	// result := DB.Where("username = ?", username).First(&user)

	// if result.Error != nil {

	// 	SendPasswordResetEmail(user.DisplayName, req.Email, code)
	// }

	// always return success, to avoid letting people fish for accounts
	return c.String(http.StatusOK, "Success!")
}

type InviteCreator struct {
	Email       string `json:"email"`
	PodSelector string `json:"podSelector"`
}

func SendInvite(c echo.Context) error {
	user_id, err := GetUserIdFromToken(c)
	if err != nil {
		return AbstractError(c)
	}

	req := InviteCreator{}
	defer c.Request().Body.Close()
	err = json.NewDecoder(c.Request().Body).Decode(&req)

	if err != nil {
		return AbstractError(c)
	}

	// get pod
	pod := Pod{}
	result := DB.Where("selector = ?", req.PodSelector).First(&pod)

	if result.Error != nil {
		return AbstractError(c)
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
		return AbstractError(c)
	}

	user := User{}

	result = DB.First(&user, user_id)

	if result.Error != nil {
		return AbstractError(c)
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

func SendPasswordResetEmail(senderName string, email string, inviteCode string) {
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

func SendRefundLimitEmail(pod Pod) {

}
