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
var SENDGRID_REFUND_LIMIT_TEMPLATE = "d-5ee1dfa663414d14b194eb770b8b65ef"
var SENDGRID_PW_RESET_TEMPLATE = "d-f05cdb7f762e48d0a1188f3b0173163d"

type InviteCreator struct {
	Email       string `json:"email"`
	PodSelector string `json:"podSelector"`
}

type DynamicData struct {
	Key   string `json:"key"`
	Value string `json:"value"`
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

	dd := []DynamicData{}

	dd = append(dd, DynamicData{
		Key:   "inviteCode",
		Value: code,
	})

	dd = append(dd, DynamicData{
		Key:   "senderName",
		Value: user.DisplayName,
	})

	emails := []string{req.Email}

	SendEmail("invite", SENDGRID_INVITE_TEMPLATE, emails, dd)

	return c.String(http.StatusOK, "Success!")
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

	dd := []DynamicData{}

	dd = append(dd, DynamicData{
		Key:   "resetCode",
		Value: code,
	})

	emails := []string{username}

	SendEmail("request", SENDGRID_PW_RESET_TEMPLATE, emails, dd)
	return c.String(http.StatusOK, "Success!")
}

func SendRefundLimitEmail(collaborators []Collaborator) {
	log.Println("SendRefundLimitEmail")

	dd := []DynamicData{}

	// list refunds that were cancelled
	// dd = append(dd, DynamicData{
	// 	Key:   "verificationCode",
	// 	Value: verificationCode.Code,
	// })

	emails := []string{}
	// email all collaborators
	for _, clbrtr := range collaborators {
		user := clbrtr.User
		emails = append(emails, user.Username)
	}

	SendEmail("refunds", SENDGRID_REFUND_LIMIT_TEMPLATE, emails, dd)
}

func SendVerificationEmail(c echo.Context) error {
	log.Println("SendVerificationEmail")
	user_id, err := GetUserIdFromToken(c)
	if err != nil {
		return AbstractError(c, "Something went wrong")
	}

	user := User{}
	result := DB.Find(&user, user_id)
	if result.Error != nil {
		return AbstractError(c, "Couldn't find user")
	}

	verificationCode := EmailVerificationCode{}

	result = DB.Where("user_id = ?", user.ID).First(&verificationCode)
	if result.Error != nil {
		verificationCode = EmailVerificationCode{
			UserID: user.ID,
			Code:   MakeInviteCode(),
		}
		DB.Create(&verificationCode)
	}

	dd := []DynamicData{}

	dd = append(dd, DynamicData{
		Key:   "verificationCode",
		Value: verificationCode.Code,
	})

	emails := []string{user.Username}

	SendEmail("verification", SENDGRID_VERIFICATION_TEMPLATE, emails, dd)

	return c.String(http.StatusOK, "Sent verification email to "+user.Username)
}

// general function used by all email routes
func SendEmail(sender string, templateId string, toEmails []string, dynamicData []DynamicData) {

	m := mail.NewV3Mail()

	address := sender + "@jamwallet.com"
	name := "Jam Wallet"
	e := mail.NewEmail(name, address)
	m.SetFrom(e)

	m.SetTemplateID(templateId)

	p := mail.NewPersonalization()
	tos := []*mail.Email{}

	for _, em := range toEmails {
		tos = append(tos, mail.NewEmail("", em))
	}

	p.AddTos(tos...)

	for _, dd := range dynamicData {
		p.SetDynamicTemplateData(dd.Key, dd.Value)
	}

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

func Direct_SendVerificationEmail(user User) {
	verificationCode := EmailVerificationCode{}

	result := DB.Where("user_id = ?", user.ID).First(&verificationCode)
	if result.Error != nil {
		verificationCode = EmailVerificationCode{
			UserID: user.ID,
			Code:   MakeInviteCode(),
		}
		DB.Create(&verificationCode)
	}

	dd := []DynamicData{}

	dd = append(dd, DynamicData{
		Key:   "verificationCode",
		Value: verificationCode.Code,
	})

	emails := []string{user.Username}

	SendEmail("verification", SENDGRID_VERIFICATION_TEMPLATE, emails, dd)
}
