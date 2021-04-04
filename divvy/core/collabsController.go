package core

import (
	"encoding/json"
	"log"
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

func CreateDivvyPod(c echo.Context) error {
	user_id, err := GetUserIdFromToken(c)

	if err != nil {
		return AbstractError(c)
	}

	req := DivvyPod{}
	defer c.Request().Body.Close()
	err = json.NewDecoder(c.Request().Body).Decode(&req)

	if err != nil {
		return AbstractError(c)
	}

	divvyPod := DivvyPod{
		UserId:      user_id,
		Name:        req.Name,
		Description: req.Description,
		Selector:    MakeSelector(DIVVY_POD_TABLE),
	}

	result := DB.Create(&divvyPod) // pass pointer of data to Create

	if result.Error != nil {
		return AbstractError(c)
	}
	// create admin collaborator
	collaborator := Collaborator{
		DivvyPodId: divvyPod.ID,
		UserId:     user_id,
		IsAdmin:    1,
		Selector:   MakeSelector(COLLABORATOR_TABLE),
	}

	result = DB.Create(&collaborator) // pass pointer of data to Create

	if result.Error != nil {
		return AbstractError(c)
	}

	return c.String(http.StatusOK, "Success!")
}

func GetDivvyPodList(c echo.Context) error {
	// get user_id from jwt
	user_id, err := GetUserIdFromToken(c)
	if err != nil {
		return AbstractError(c)
	}

	// get all my collaborator records
	collaborators := []Collaborator{}
	result := DB.Where("user_id = ?", user_id).Find(&collaborators)
	if result.Error != nil {
		return AbstractError(c)
	}
	// take collab_id from all results

	// use collaborator records to pull my collabs
	divvyPods := []DivvyPod{}

	podIds := []int{1, 2, 3}
	// result = DB.Where("id = ?", user_id).Find(&collabs)
	result = DB.Find(&divvyPods, podIds)
	// SELECT * FROM collabs WHERE id IN (1,2,3);

	log.Println("got divvyPods")
	log.Println(divvyPods)

	if result.Error != nil {
		return AbstractError(c)
	}

	return c.JSON(http.StatusOK, divvyPods)
}

func GetDivvyPod(c echo.Context) error {
	// get user_id from jwt
	user_id, err := GetUserIdFromToken(c)

	if err != nil {
		return AbstractError(c)
	}
	// get from params
	selector := c.Param("selector")

	divvyPod := DivvyPod{}
	result := DB.Where("selector = ?", selector).First(&divvyPod)
	if result.Error != nil {
		return AbstractError(c)
	}

	// make sure this user is part of the collab
	collaborator := Collaborator{}
	result = DB.Where("user_id = ?", user_id).Where("divvy_pod_id = ?", divvyPod.ID).First(&collaborator)

	if result.Error != nil {
		return AbstractError(c)
	}

	return c.JSON(http.StatusOK, divvyPod)
}
