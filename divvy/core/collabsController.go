package core

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
)

type CollabsResponse struct {
	Collabs []CollabResponse `json:"collabs"`
}

type CollabResponse struct {
	Name     string `json:"name"`
	Selector string `json:"selector"`
}

func CreateCollab(c echo.Context) error {
	user_id, err := GetUserIdFromToken(c)

	if err != nil {
		return AbstractError(c)
	}

	req := Collab{}
	defer c.Request().Body.Close()
	err = json.NewDecoder(c.Request().Body).Decode(&req)

	if err != nil {
		return AbstractError(c)
	}

	collab := Collab{
		UserId:      user_id,
		Name:        req.Name,
		Description: req.Description,
	}

	result := DB.Create(&collab) // pass pointer of data to Create

	if result.Error != nil {
		return AbstractError(c)
	}

	// create admin collaborator
	collaborator := Collaborator{
		CollabId: collab.ID,
		UserId:   user_id,
		IsAdmin:  1,
		Selector: MakeSelector(COLLABORATOR_TABLE),
	}

	result = DB.Create(&collaborator) // pass pointer of data to Create

	if result.Error != nil {
		return AbstractError(c)
	}

	return c.String(http.StatusOK, "Success!")
}

func GetCollabs(c echo.Context) error {
	// get user_id from jwt
	user_id, err := GetUserIdFromToken(c)
	if err != nil {
		return AbstractError(c)
	}

	// find my collabs
	// get all my collaborator records
	collaborators := []Collaborator{}
	result := DB.Where("user_id = ?", user_id).Find(&collaborators)
	if result.Error != nil {
		return AbstractError(c)
	}

	collabIds := []int{1, 2, 3}

	collabs := []Collab{}
	result = DB.Where("id = ?", user_id).Find(&collaborators)
	DB.Find(&collabs, collabIds)
	// SELECT * FROM collabs WHERE id IN (1,2,3);

	if result.Error != nil {
		return AbstractError(c)
	}

	return c.JSON(http.StatusOK, json.NewEncoder(c.Response()).Encode(collabs))
}
