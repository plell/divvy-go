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
		UserID:      user_id,
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
		PodID:    pod.ID,
		UserID:   user_id,
		IsAdmin:  true,
		Selector: MakeSelector(COLLABORATOR_TABLE),
	}

	result = DB.Create(&collaborator) // pass pointer of data to Create

	if result.Error != nil {
		return AbstractError(c)
	}

	podResponse := PodAPI{
		Name:        pod.Name,
		Description: pod.Description,
		Selector:    pod.Selector,
	}

	return c.JSON(http.StatusOK, podResponse)
}

type Collaborators []Collaborator

func getPodIdsFromCollaborators(col Collaborators) []uint {
	var list []uint
	for _, collaborator := range col {
		podId := collaborator.PodID
		if !ContainsInt(list, podId) {
			list = append(list, podId)
		}
	}
	return list
}

func getUserIdsFromCollaborators(col Collaborators) []uint {
	var list []uint
	for _, collaborator := range col {
		userId := collaborator.UserID
		log.Println("getUserIdsFromCollaborators")
		log.Println(userId)
		if !ContainsInt(list, userId) {
			list = append(list, userId)
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

	// test sending an email
	// SendEmail()

	// get all my collaborator records
	collaborators := []Collaborator{}
	result := DB.Where("user_id = ?", user_id).Find(&collaborators)
	if result.Error != nil {
		return AbstractError(c)
	}

	pods := []Pod{}
	podIds := getPodIdsFromCollaborators(collaborators)
	// result = DB.Where("id = ?", user_id).Find(&collabs)

	if len(podIds) == 0 {
		log.Println("NO PODS!")
		return c.JSON(http.StatusOK, pods)
	}

	// IF podIds is empty it returns all!
	// SELECT * FROM divvy_pods WHERE id IN (1,2,3);
	result = DB.Where(podIds).Find(&pods)
	if result.Error != nil {
		return AbstractError(c)
	}

	podsList := []PodAPI{}

	for _, p := range pods {
		formattedPod := BuildPod(p)
		podsList = append(podsList, formattedPod)
	}

	return c.JSON(http.StatusOK, podsList)
}

type PodResponse struct {
	Pod     Pod               `json:"pod"`
	Members []CollaboratorAPI `json:"members"`
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

	response := PodResponse{
		Pod:     pod,
		Members: members,
	}

	return c.JSON(http.StatusOK, response)
}

type PodJoiner struct {
	Code string `json:"code"`
}

func JoinPod(c echo.Context) error {
	user_id, err := GetUserIdFromToken(c)
	if err != nil {
		return AbstractError(c)
	}

	req := PodJoiner{}
	defer c.Request().Body.Close()
	err = json.NewDecoder(c.Request().Body).Decode(&req)

	// get user
	user := User{}
	result := DB.First(&user, user_id)
	if result.Error != nil {
		return AbstractError(c)
	}

	if err != nil {
		return AbstractError(c)
	}

	invite := Invite{}
	result = DB.Where("code = ?", req.Code).First(&invite)
	if result.Error != nil {
		return AbstractError(c)
	}

	// if user.Username != invite.Email {
	// 	// this code was made for a different user
	// 	return AbstractError(c)
	// }

	collaborator := Collaborator{
		UserID:   user_id,
		PodID:    invite.PodID,
		IsAdmin:  false,
		Selector: MakeSelector(COLLABORATOR_TABLE),
	}

	result = DB.Create(&collaborator)

	if result.Error != nil {
		return AbstractError(c)
	}

	// delete the invite, its been used
	DB.Delete(&invite)

	return c.String(http.StatusOK, "Success")
}

func GetInvites(c echo.Context) error {
	// user_id, err := GetUserIdFromToken(c)
	// if err != nil {
	// 	return AbstractError(c)
	// }

	podSelector := c.Param("podSelector")

	pod := Pod{}
	result := DB.Where("selector = ?", podSelector).Find(&pod)
	if result.Error != nil {
		return AbstractError(c)
	}

	// get user
	invites := []InviteAPI{}
	result = DB.Model(&Invite{}).Where("pod_id = ?", pod.ID).Find(&invites)

	if result.Error != nil {
		return AbstractError(c)
	}

	return c.JSON(http.StatusOK, invites)
}

func DeleteInvite(c echo.Context) error {
	// user_id, err := GetUserIdFromToken(c)
	// if err != nil {
	// 	return AbstractError(c)
	// }

	selector := c.Param("selector")

	// get user
	invite := Invite{}

	result := DB.Where("selector = ?", selector).First(&invite)
	if result.Error != nil {
		return AbstractError(c)
	}

	DB.Delete(&invite)

	return c.String(http.StatusOK, "OK!")
}

func FindAvatarByUserId(avatars []Avatar, userId uint) Avatar {
	for _, avatar := range avatars {
		if userId == avatar.UserID {
			return avatar
		}
	}
	return Avatar{}
}
