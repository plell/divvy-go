package core

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

func CreatePod(c echo.Context) error {
	user_id, err := GetUserIdFromToken(c)

	if err != nil {
		return AbstractError(c)
	}

	req := Pod{}
	defer c.Request().Body.Close()
	err = json.NewDecoder(c.Request().Body).Decode(&req)

	if err != nil {
		return AbstractError(c)
	}

	pod := Pod{
		UserId:      user_id,
		Name:        req.Name,
		Description: req.Description,
		Selector:    MakeSelector(POD_TABLE),
	}

	result := DB.Create(&pod) // pass pointer of data to Create

	if result.Error != nil {
		return AbstractError(c)
	}
	// create admin collaborator
	collaborator := Collaborator{
		PodId:    pod.ID,
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

type Collaborators []Collaborator

func getPodIdsFromCollaborators(col Collaborators) []uint {
	var list []uint
	for _, collaborator := range col {
		podId := collaborator.PodId
		if !ContainsInt(list, podId) {
			list = append(list, podId)
		}
	}
	return list
}

func getUserIdsFromCollaborators(col Collaborators) []uint {
	var list []uint
	for _, collaborator := range col {
		podId := collaborator.PodId
		if !ContainsInt(list, podId) {
			list = append(list, podId)
		}
	}
	return list
}

func GetPodList(c echo.Context) error {
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

	pods := []PodAPI{}
	podIds := getPodIdsFromCollaborators(collaborators)
	// result = DB.Where("id = ?", user_id).Find(&collabs)

	if len(podIds) == 0 {
		log.Println("NO PODS!")
		return c.JSON(http.StatusOK, pods)
	}
	// IF podIds is empty is returns all!
	// SELECT * FROM divvy_pods WHERE id IN (1,2,3);
	result = DB.Model(&Pod{}).Where(podIds).Find(&pods)

	if result.Error != nil {
		return AbstractError(c)
	}

	return c.JSON(http.StatusOK, pods)
}

type PodResponse struct {
	Pod     Pod       `json:"pod"`
	Members []UserAPI `json:"members"`
}

func GetPod(c echo.Context) error {
	// get user_id from jwt
	user_id, err := GetUserIdFromToken(c)

	if err != nil {
		return AbstractError(c)
	}
	// get from params
	selector := c.Param("selector")

	pod := Pod{}
	result := DB.Where("selector = ?", selector).First(&pod)
	if result.Error != nil {
		return AbstractError(c)
	}

	// make sure this user is part of the collab
	collaborator := Collaborator{}
	result = DB.Where("user_id = ?", user_id).Where("pod_id = ?", pod.ID).First(&collaborator)

	if result.Error != nil {
		return AbstractError(c)
	}

	// get collaborators
	collaborators := []Collaborator{}
	// SELECT * FROM divvy_pods WHERE id IN (1,2,3);
	result = DB.Model(&Collaborator{}).Where("pod_id = ?", pod.ID).Find(&collaborators)

	if result.Error != nil {
		return AbstractError(c)
	}

	// get all my user records
	users := []User{}
	userIds := getUserIdsFromCollaborators(collaborators)

	result = DB.Where(userIds).Find(&users)

	if result.Error != nil {
		return AbstractError(c)
	}

	// get all my avatar records
	avatars := []Avatar{}
	result = DB.Where("user_id = ?", userIds).Find(&avatars)
	if result.Error != nil {
		return AbstractError(c)
	}

	members := []UserAPI{}
	// append avatars to users
	for _, u := range users {
		thisAvatar := FindAvatarByUserId(avatars, u.ID)
		user := BuildUser(u, thisAvatar)
		members = append(members, user)
	}

	response := PodResponse{
		Pod:     pod,
		Members: members,
	}

	return c.JSON(http.StatusOK, response)
}

func FindAvatarByUserId(avatars []Avatar, userId uint) Avatar {
	for _, avatar := range avatars {
		if userId == avatar.UserId {
			return avatar
		}
	}
	return Avatar{}
}