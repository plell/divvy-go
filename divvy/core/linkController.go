package core

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

func CreateCheckoutLink(c echo.Context) error {
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

	if request.Amount < 100 {
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
	}

	log.Println("THE LINK")
	log.Println(link)

	result = DB.Create(&link)
	if result.Error != nil {
		return c.String(http.StatusInternalServerError, "couldn't create link")
	}

	return c.JSON(http.StatusOK, link)
}
