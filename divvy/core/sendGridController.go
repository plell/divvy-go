// using SendGrid's Go Library
// https://github.com/sendgrid/sendgrid-go
package core

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/stripe/stripe-go/v72"
)

var SENDGRID_INVITE_TEMPLATE = "d-4416cc09847445c9867d1e9d3cf09dcc"
var SENDGRID_VERIFICATION_TEMPLATE = "d-c0ae63959f8c4a30aca48f6599b07ed4"
var SENDGRID_REFUND_LIMIT_TEMPLATE = "d-5ee1dfa663414d14b194eb770b8b65ef"
var SENDGRID_PW_RESET_TEMPLATE = "d-f05cdb7f762e48d0a1188f3b0173163d"
var SENDGRID_BETA_INVITE_TEMPLATE = "d-98ebea08ca684fde9c27aebf641680da"
var SENDGRID_BETA_INVITE_REQUEST_CONFIRMATION_TEMPLATE = "d-81e126e7cf1549069b608cab85608c2c"

var SENDGRID_BETA_INVITE_REQUEST_APPROVAL_TEMPLATE = "d-cc8f578b0cd8481f81232bfd8b467a90"

var SENDGRID_REFUND_SCHEDULED_TEMPLATE = "d-18dc29d5d64f4c20ac09a6d9ca4aeea2"
var SENDGRID_REFUND_CANCELLED_TEMPLATE = "d-d376ac9bf2194d53940f156308215b6f"
var SENDGRID_WALLET_UPDATED_TEMPLATE = "d-8d0ae75cf44d42f08b3142725f18f45c"
var SENDGRID_WALLET_JOINED_TEMPLATE = "d-6fdd284b1b054066828410fefa25c94b"
var SENDGRID_WALLET_DESTROYED_TEMPLATE = "d-a72b5e54002a4b50a055f68862bea2f8"

var SENDGRID_PAYOUT_TEMPLATE = "d-a513fb876a094b90920e5159d08734fb"
var SENDGRID_PAYMENT_RECEIVED_TEMPLATE = "d-5a26312bde7d41f8a292a654f9a60e8e"

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

func SendBetaInvite(c echo.Context) error {

	req := InviteCreator{}
	defer c.Request().Body.Close()
	err := json.NewDecoder(c.Request().Body).Decode(&req)

	if err != nil {
		return AbstractError(c, "Couldn't decode request")
	}

	code := MakeInviteCode()

	betaKey := BetaKey{
		BetaKey: code,
	}

	result := DB.Create(&betaKey)

	if result.Error != nil {
		return AbstractError(c, "Couldn't create beta key")
	}

	dd := []DynamicData{}

	dd = append(dd, DynamicData{
		Key:   "betaKey",
		Value: code,
	})

	emails := []string{req.Email}

	SendEmail("invite", SENDGRID_BETA_INVITE_TEMPLATE, emails, dd)

	return c.String(http.StatusOK, "Success!")
}

func SendBetaInviteRequest(c echo.Context) error {

	req := BetaKeyRequest{}
	defer c.Request().Body.Close()
	err := json.NewDecoder(c.Request().Body).Decode(&req)

	if err != nil {
		return AbstractError(c, "Couldn't decode request")
	}

	code := ""
	betaKeyRequest := BetaKeyRequest{}
	// make sure its not a duplicate
	result := DB.Where("email = ?", req.Email).First(&betaKeyRequest)

	// if not found
	if result.Error != nil {

		code = MakeSelector(BETA_KEY_REQUESTS_TABLE)

		betaKeyRequest := BetaKeyRequest{
			AcceptCode: code,
			Email:      req.Email,
			Message:    req.Message,
		}

		result = DB.Create(&betaKeyRequest)

		if result.Error != nil {
			return AbstractError(c, "Couldn't create beta key request")
		}
	} else {
		code = betaKeyRequest.AcceptCode
	}

	dd := []DynamicData{}

	emails := []string{req.Email}

	SendEmail("beta", SENDGRID_BETA_INVITE_REQUEST_CONFIRMATION_TEMPLATE, emails, dd)

	dd = []DynamicData{}
	dd = append(dd, DynamicData{
		Key:   "acceptCode",
		Value: code,
	})

	dd = []DynamicData{}
	dd = append(dd, DynamicData{
		Key:   "message",
		Value: req.Message,
	})

	url := CLIENT_URL + "?betaAcceptCode=" + code
	dd = append(dd, DynamicData{
		Key:   "url",
		Value: url,
	})

	dd = append(dd, DynamicData{
		Key:   "email",
		Value: req.Email,
	})

	super_admin_email := os.Getenv("SUPER_ADMIN_EMAIL")
	emails = []string{super_admin_email}

	SendEmail("admin", SENDGRID_BETA_INVITE_REQUEST_APPROVAL_TEMPLATE, emails, dd)

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

func SendRefundLimitEmail(podSelector string) {
	log.Println("SendRefundLimitEmail")

	collaborators, podName, err := getCollaboratorsFromPodSelector(podSelector)
	if err != nil {
		return
	}
	// list refunds that were cancelled
	dd := []DynamicData{}
	dd = append(dd, DynamicData{
		Key:   "walletName",
		Value: podName,
	})

	emails := []string{}
	// email all collaborators
	for _, clbrtr := range collaborators {
		user := clbrtr.User
		emails = append(emails, user.Username)
	}

	SendEmail("refunds", SENDGRID_REFUND_LIMIT_TEMPLATE, emails, dd)
}

func getCollaboratorsFromPodSelector(podSelector string) ([]Collaborator, string, error) {
	var throwerror error = nil
	pod := Pod{}
	collaborators := []Collaborator{}

	result := DB.Where("selector = ?", podSelector).First(&pod)
	if result.Error != nil {
		throwerror = nil
	}

	result = DB.Preload("User").Where("pod_id = ?", pod.ID).Find(&collaborators)
	if result.Error != nil {
		throwerror = nil
	}

	podName := pod.Name

	return collaborators, podName, throwerror
}

func getCollaboratorsFromPodID(podID uint) ([]Collaborator, string, error) {
	var throwerror error = nil
	pod := Pod{}
	collaborators := []Collaborator{}

	result := DB.First(&pod, podID)
	if result.Error != nil {
		throwerror = nil
	}

	result = DB.Preload("User").Where("pod_id = ?", podID).Find(&collaborators)
	if result.Error != nil {
		throwerror = nil
	}

	podName := pod.Name

	return collaborators, podName, throwerror
}

func SendRefundScheduledEmail(podSelector string) {
	log.Println("SendRefundScheduledEmail")

	collaborators, podName, err := getCollaboratorsFromPodSelector(podSelector)
	if err != nil {
		return
	}

	dd := []DynamicData{}
	dd = append(dd, DynamicData{
		Key:   "walletName",
		Value: podName,
	})
	emails := []string{}
	// email all collaborators
	for _, clbrtr := range collaborators {
		user := clbrtr.User
		emails = append(emails, user.Username)
	}

	SendEmail("refunds", SENDGRID_REFUND_SCHEDULED_TEMPLATE, emails, dd)
}

func SendRefundCancelledEmail(podSelector string) {
	log.Println("SendRefundScheduledEmail")

	collaborators, podName, err := getCollaboratorsFromPodSelector(podSelector)
	if err != nil {
		return
	}
	dd := []DynamicData{}
	dd = append(dd, DynamicData{
		Key:   "walletName",
		Value: podName,
	})
	emails := []string{}
	// email all collaborators
	for _, clbrtr := range collaborators {
		user := clbrtr.User
		emails = append(emails, user.Username)
	}

	SendEmail("refunds", SENDGRID_REFUND_CANCELLED_TEMPLATE, emails, dd)
}

// Amount           int64  `json:"amount"`
// TransactionCount int    `json:"transactionCount"`
// Fees             int64  `json:"fees"`
// Email            string `json:"email"`
// UserID           uint   `json:"userId"`

func SendPayoutEmail(payout UserPayout) {
	log.Println("SendPayoutEmail")

	dd := []DynamicData{}
	dd = append(dd, DynamicData{
		Key:   "amount",
		Value: FormatAmountToString(payout.Amount),
	})
	dd = append(dd, DynamicData{
		Key:   "fees",
		Value: FormatAmountToString(payout.Fees),
	})
	dd = append(dd, DynamicData{
		Key:   "transactionCount",
		Value: strconv.Itoa(payout.TransactionCount),
	})

	emails := []string{}
	emails = append(emails, payout.Email)

	SendEmail("payouts", SENDGRID_PAYOUT_TEMPLATE, emails, dd)
}

func SendPaymentReceivedEmail(c stripe.Charge) {
	log.Println("SendPaymentReceivedEmail")

	podSelector := ""
	amountAfterFees := 0
	if _, ok := c.Metadata["podSelector"]; ok {
		podSelector = c.Metadata["podSelector"]
	}

	if _, ok := c.Metadata["amountAfterFees"]; ok {
		amountAfterFees, _ = strconv.Atoi(c.Metadata["amountAfterFees"])
	}

	collaborators, podName, err := getCollaboratorsFromPodSelector(podSelector)
	if err != nil {
		return
	}

	dd := []DynamicData{}
	dd = append(dd, DynamicData{
		Key:   "amount",
		Value: FormatAmountToString(c.Amount),
	})
	dd = append(dd, DynamicData{
		Key:   "amountAfterFees",
		Value: FormatAmountToString(int64(amountAfterFees)),
	})
	dd = append(dd, DynamicData{
		Key:   "podName",
		Value: podName,
	})

	emails := []string{}
	// email all collaborators
	for _, clbrtr := range collaborators {
		user := clbrtr.User
		emails = append(emails, user.Username)
	}

	SendEmail("payment", SENDGRID_PAYMENT_RECEIVED_TEMPLATE, emails, dd)
}

func SendWalletUpdatedEmail(podSelector string) {
	log.Println("SendRefundScheduledEmail")

	collaborators, podName, err := getCollaboratorsFromPodSelector(podSelector)
	if err != nil {
		return
	}
	dd := []DynamicData{}
	dd = append(dd, DynamicData{
		Key:   "walletName",
		Value: podName,
	})
	emails := []string{}
	// email all collaborators
	for _, clbrtr := range collaborators {
		user := clbrtr.User
		emails = append(emails, user.Username)
	}

	SendEmail("notification", SENDGRID_WALLET_UPDATED_TEMPLATE, emails, dd)
}

func SendWalletJoinedEmail(podID uint) {
	log.Println("SendRefundScheduledEmail")

	collaborators, podName, err := getCollaboratorsFromPodID(podID)
	if err != nil {
		return
	}
	dd := []DynamicData{}
	dd = append(dd, DynamicData{
		Key:   "walletName",
		Value: podName,
	})
	emails := []string{}
	// email all collaborators
	for _, clbrtr := range collaborators {
		user := clbrtr.User
		emails = append(emails, user.Username)
	}

	SendEmail("notification", SENDGRID_WALLET_JOINED_TEMPLATE, emails, dd)
}

func SendWalletDestroyedEmail(podSelector string) {
	log.Println("SendRefundScheduledEmail")

	collaborators, podName, err := getCollaboratorsFromPodSelector(podSelector)
	if err != nil {
		return
	}
	dd := []DynamicData{}
	dd = append(dd, DynamicData{
		Key:   "walletName",
		Value: podName,
	})
	emails := []string{}
	// email all collaborators
	for _, clbrtr := range collaborators {
		user := clbrtr.User
		emails = append(emails, user.Username)
	}

	SendEmail("notification", SENDGRID_WALLET_DESTROYED_TEMPLATE, emails, dd)
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

// send email to all collaborators
