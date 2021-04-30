package core

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

type CollaboratorRoleRequest struct {
	RoleTypeID uint   `json:"roleTypeId"`
	Selector   string `json:"selector"`
}

func UpdateCollaboratorRole(c echo.Context) error {

	req := CollaboratorRoleRequest{}
	defer c.Request().Body.Close()
	err := json.NewDecoder(c.Request().Body).Decode(&req)

	if err != nil {
		AbstractError(c, "Something went wrong")
		return AbstractError(c, "Something went wrong")
	}

	// get request collaborator
	collaborator := Collaborator{}
	result := DB.Where("selector = ?", req.Selector).First(&collaborator)
	if result.Error != nil {
		return AbstractError(c, "Couldn't find collaborator")
	}

	collaborator.RoleTypeID = req.RoleTypeID

	result = DB.Save(&collaborator)
	if result.Error != nil {
		return AbstractError(c, "Couldn't update!")
	}

	return c.String(http.StatusOK, "Success")
}

func DeleteCollaborator(c echo.Context) error {
	// user_id, err := GetUserIdFromToken(c)
	// if err != nil {
	// 	return AbstractError(c,"Something went wrong")
	// }

	selector := c.Param("selector")

	log.Println("Selector")
	log.Println(selector)
	collaborator := Collaborator{}

	result := DB.Where("selector = ?", selector).First(&collaborator)
	if result.Error != nil {
		return AbstractError(c, "Something went wrong")
	}

	// make sure pod has more than 1 collaborator
	pod := Pod{}
	result = DB.Preload("Collaborators").First(&pod, collaborator.ID)
	if result.Error != nil {
		return AbstractError(c, "Something went wrong")
	}

	if len(pod.Collaborators) < 2 {
		return AbstractError(c, "You can't leave a wallet when you're the only member.")
	}

	result = DB.Delete(&collaborator)
	if result.Error != nil {
		return AbstractError(c, "Something went wrong")
	}

	return c.String(http.StatusOK, "Success")
}

func GetCollaboratorList(c echo.Context) error {
	// user_id, err := GetUserIdFromToken(c)
	// if err != nil {
	// 	return AbstractError(c,"Something went wrong")
	// }

	podSelector := c.Param("podSelector")

	// get pod
	pod := Pod{}
	result := DB.Where("selector = ?", podSelector).Find(&pod)

	// get collaborators
	collaborators := []Collaborator{}
	result = DB.Preload("User").Preload("User.Avatar").Where("pod_id = ?", pod.ID).Find(&collaborators)

	if result.Error != nil {
		return AbstractError(c, "Something went wrong")
	}

	members := []CollaboratorAPI{}
	// append avatars to users
	for _, c := range collaborators {
		user := BuildUserFromCollaborator(c)
		members = append(members, user)
	}

	return c.JSON(http.StatusOK, members)
}
