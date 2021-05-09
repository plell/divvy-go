package core

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
)

func GetBetaRequestList(c echo.Context) error {

	betaKeyRequests := []BetaKeyRequest{}
	result := DB.Find(&betaKeyRequests)
	if result.Error != nil {
		return AbstractError(c, "Couldn't get beta key requests")
	}

	return c.JSON(http.StatusOK, betaKeyRequests)
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

	// delete beta request if exists
	betaKeyRequest := BetaKeyRequest{}
	result = DB.Where("email = ?", req.Email).First(&betaKeyRequest)
	if result.Error == nil {
		result = DB.Delete(&betaKeyRequest)
		if result.Error != nil {
			return AbstractError(c, "Couldn't delete beta key request")
		}
	}

	return c.String(http.StatusOK, "Success!")
}

func SendBetaInviteRequest(c echo.Context) error {

	req := BetaKeyRequest{}
	defer c.Request().Body.Close()
	err := json.NewDecoder(c.Request().Body).Decode(&req)

	if err != nil {
		return AbstractError(c, "Couldn't decode request")
	}

	betaKeyRequest := BetaKeyRequest{}
	// make sure its not a duplicate
	result := DB.Where("email = ?", req.Email).First(&betaKeyRequest)

	// if not found
	if result.Error != nil {

		betaKeyRequest := BetaKeyRequest{
			Email:   req.Email,
			Message: req.Message,
		}

		result = DB.Create(&betaKeyRequest)

		if result.Error != nil {
			return AbstractError(c, "Couldn't create beta key request")
		}
	}

	dd := []DynamicData{}

	emails := []string{req.Email}

	SendEmail("beta", SENDGRID_BETA_INVITE_REQUEST_CONFIRMATION_TEMPLATE, emails, dd)

	activeBetaKeyRequests := []BetaKeyRequest{}
	// make sure its not a duplicate
	result = DB.Find(&activeBetaKeyRequests)
	if result.Error != nil {
		return AbstractError(c, "Couldn't get beta key requests")
	}

	dd = []DynamicData{}

	dd = []DynamicData{}
	dd = append(dd, DynamicData{
		Key:   "message",
		Value: req.Message,
	})

	dd = append(dd, DynamicData{
		Key:   "email",
		Value: req.Email,
	})

	dd = append(dd, DynamicData{
		Key:   "activeRequestCount",
		Value: strconv.Itoa(len(activeBetaKeyRequests)),
	})

	super_admin_email := os.Getenv("SUPER_ADMIN_EMAIL")
	emails = []string{super_admin_email}

	SendEmail("admin", SENDGRID_BETA_INVITE_REQUEST_APPROVAL_TEMPLATE, emails, dd)

	return c.String(http.StatusOK, "Success!")
}
