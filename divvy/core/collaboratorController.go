package core

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

type CollaboratorRoleRequest struct {
	IsAdmin  bool   `json:"isAdmin"`
	Selector string `json:"selector"`
}

func UpdateCollaboratorAdmin(c echo.Context) error {
	// user_id, err := GetUserIdFromToken(c)
	// if err != nil {
	// 	return AbstractError(c)
	// }

	req := CollaboratorRoleRequest{}
	defer c.Request().Body.Close()
	err := json.NewDecoder(c.Request().Body).Decode(&req)

	if err != nil {
		return AbstractError(c)
	}

	// get request collaborator
	collaborator := Collaborator{}
	result := DB.Where("selector = ?", req.Selector).First(&collaborator)
	if result.Error != nil {
		return AbstractError(c)
	}

	log.Println("req")
	log.Println(req)

	// collaborator.IsAdmin = req.IsAdmin
	DB.Model(&collaborator).Update("is_admin", req.IsAdmin)

	// DB.Save(&collaborator)

	return c.String(http.StatusOK, "Success")
}

func DeleteCollaborator(c echo.Context) error {
	// user_id, err := GetUserIdFromToken(c)
	// if err != nil {
	// 	return AbstractError(c)
	// }

	selector := c.Param("selector")

	log.Println("Selector")
	log.Println(selector)
	collaborator := Collaborator{}

	result := DB.Where("selector = ?", selector).Delete(&collaborator)
	if result.Error != nil {
		return AbstractError(c)
	}

	return c.String(http.StatusOK, "Success")
}

func GetCollaboratorList(c echo.Context) error {
	// user_id, err := GetUserIdFromToken(c)
	// if err != nil {
	// 	return AbstractError(c)
	// }

	podSelector := c.Param("podSelector")

	// get pod
	pod := Pod{}
	result := DB.Where("selector = ?", podSelector).Find(&pod)

	// get collaborators
	collaborators := []Collaborator{}
	result = DB.Preload("User").Preload("User.Avatar").Where("pod_id = ?", pod.ID).Find(&collaborators)

	if result.Error != nil {
		return AbstractError(c)
	}

	members := []CollaboratorAPI{}
	// append avatars to users
	for _, c := range collaborators {
		user := BuildUserFromCollaborator(c)
		members = append(members, user)
	}

	return c.JSON(http.StatusOK, members)
}
