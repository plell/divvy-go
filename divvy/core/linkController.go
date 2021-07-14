package core

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

func CreateCheckoutLink(c echo.Context) error {
	isFixedAmount := true

	user_id, err := GetUserIdFromToken(c)
	if err != nil {
		return AbstractError(c, "Something went wrong")
	}

	// here decode the pod selector and include it in TRANSFER GROUP
	request := Link{}
	defer c.Request().Body.Close()
	err = json.NewDecoder(c.Request().Body).Decode(&request)
	if err != nil {
		return c.String(http.StatusInternalServerError, "can't decode request")
	}

	if request.Amount == 0 {
		// this is a variable donation, donors choose the price
		isFixedAmount = false
	} else if request.Amount < 100 {
		return c.String(http.StatusInternalServerError, "Amount minimum is 1USD")
	}

	// get pod for metadata
	pod := Pod{}
	result := DB.Where("selector = ?", request.PodSelector).First(&pod)
	if result.Error != nil {
		return c.String(http.StatusInternalServerError, "no pod")
	}

	coll := Collaborator{}
	result = DB.Preload("User").Where("user_id = ?", user_id).Where("pod_id = ?", pod.ID).First(&coll)
	if result.Error != nil {
		return c.String(http.StatusInternalServerError, "no user")
	}

	link := Link{
		Amount:               request.Amount,
		PodSelector:          pod.Selector,
		UserSelector:         coll.User.Selector,
		CollaboratorSelector: coll.Selector,
		Selector:             MakeSelector(LINK_TABLE),
		IsFixedAmount:        isFixedAmount,
	}

	result = DB.Create(&link)
	if result.Error != nil {
		return c.String(http.StatusInternalServerError, "couldn't create link")
	}

	return c.JSON(http.StatusOK, link)
}

func GetCheckoutLink(c echo.Context) error {
	selector := c.Param("selector")
	link := Link{}

	result := DB.Where("selector = ?", selector).First(&link)
	if result.Error != nil {
		return c.String(http.StatusInternalServerError, "couldn't get link")
	}

	// get pod for metadata
	pod := Pod{}
	result = DB.Where("selector = ?", link.PodSelector).First(&pod)
	if result.Error != nil {
		return c.String(http.StatusInternalServerError, "no pod")
	}

	// add podname to link for ui display
	link.Name = pod.Name

	return c.JSON(http.StatusOK, link)
}
