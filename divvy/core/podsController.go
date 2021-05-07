package core

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func CreatePod(c echo.Context) error {
	user_id, err := GetUserIdFromToken(c)

	if err != nil {
		return AbstractError(c, "Something went wrong")
	}

	req := Pod{}
	defer c.Request().Body.Close()
	err = json.NewDecoder(c.Request().Body).Decode(&req)

	if err != nil {
		return AbstractError(c, "Something went wrong")
	}

	pod := Pod{
		UserID:          user_id,
		Name:            req.Name,
		Description:     req.Description,
		Selector:        MakeSelector(POD_TABLE),
		PayoutTypeId:    req.PayoutTypeId,
		LifecycleTypeId: req.LifecycleTypeId,
	}

	result := DB.Create(&pod) // pass pointer of data to Create

	if result.Error != nil {
		return AbstractError(c, "Something went wrong")
	}

	// make pod rules
	// rules := []PodRule{{
	// 	Value:         "6",
	// 	PodRuleTypeID: POD_RULE_MAX_GROUP_SIZE,
	// 	PodID:         pod.ID,
	// }}
	// result = DB.Create(&rules) // pass pointer of data to Create
	// if result.Error != nil {
	// 	return AbstractError(c, "Something went wrong")
	// }

	// create admin collaborator
	collaborator := Collaborator{
		PodID:      pod.ID,
		UserID:     user_id,
		RoleTypeID: ROLE_TYPE_ADMIN,
		Selector:   MakeSelector(COLLABORATOR_TABLE),
	}

	result = DB.Create(&collaborator) // pass pointer of data to Create
	if result.Error != nil {
		return AbstractError(c, "Something went wrong")
	}

	podResponse := PodAPI{
		Name:            pod.Name,
		Description:     pod.Description,
		Selector:        pod.Selector,
		PayoutTypeId:    pod.PayoutTypeId,
		LifecycleTypeId: pod.LifecycleTypeId,
	}

	return c.JSON(http.StatusOK, podResponse)
}

func UpdatePod(c echo.Context) error {
	user_id, err := GetUserIdFromToken(c)
	if err != nil {
		return AbstractError(c, "Something went wrong")
	}

	req := Pod{}
	defer c.Request().Body.Close()
	err = json.NewDecoder(c.Request().Body).Decode(&req)

	if err != nil {
		return AbstractError(c, "Something went wrong")
	}

	// find pod

	podSelector := c.Param("podSelector")
	pod := Pod{}
	result := DB.Where("selector = ?", podSelector).First(&pod) // pass pointer of data to Create
	if result.Error != nil {
		return AbstractError(c, "Couldn't find pod")
	}

	pod.UserID = user_id
	pod.Name = req.Name
	pod.Description = req.Description
	pod.PayoutTypeId = req.PayoutTypeId
	pod.LifecycleTypeId = req.LifecycleTypeId

	result = DB.Save(&pod) // pass pointer of data to Create

	if result.Error != nil {
		return AbstractError(c, "Couldn't save pod")
	}

	// make pod rules
	// rules := []PodRule{{
	// 	Value:         "6",
	// 	PodRuleTypeID: POD_RULE_MAX_GROUP_SIZE,
	// 	PodID:         pod.ID,
	// }}
	// result = DB.Create(&rules) // pass pointer of data to Create
	// if result.Error != nil {
	// 	return AbstractError(c, "Something went wrong")
	// }

	podResponse := PodAPI{
		Name:            pod.Name,
		Description:     pod.Description,
		Selector:        pod.Selector,
		PayoutTypeId:    pod.PayoutTypeId,
		LifecycleTypeId: pod.LifecycleTypeId,
	}

	// send email to all collaborators
	SendWalletUpdatedEmail(pod.Selector)

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
		return AbstractError(c, "Something went wrong")
	}

	// test sending an email
	// SendEmail()

	// get all my collaborator records
	collaborators := []Collaborator{}
	result := DB.Where("user_id = ?", user_id).Find(&collaborators)
	if result.Error != nil {
		return AbstractError(c, "Something went wrong")
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
	result = DB.Preload("LifecycleType").Preload("PayoutType").Preload("Collaborators").Where(podIds).Find(&pods)
	if result.Error != nil {
		return AbstractError(c, "Something went wrong")
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
	podSelector := c.Param("podSelector")

	pod := Pod{}
	result := DB.Preload("PayoutType").Preload("LifecycleType").Where("selector = ?", podSelector).First(&pod)
	if result.Error != nil {
		return AbstractError(c, "Something went wrong")
	}

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
		return AbstractError(c, "Couldn't read token")
	}

	req := PodJoiner{}
	defer c.Request().Body.Close()
	err = json.NewDecoder(c.Request().Body).Decode(&req)
	if err != nil {
		return AbstractError(c, "Couldn't read request")
	}
	// get user
	user := User{}
	result := DB.First(&user, user_id)
	if result.Error != nil {
		return AbstractError(c, "Couldn't find user")
	}

	invite := Invite{}
	result = DB.Where("code = ?", req.Code).First(&invite)
	if result.Error != nil {
		return AbstractError(c, "This code is not valid")
	}

	// does the collaborator already exist?
	coll := Collaborator{}
	result = DB.Where("user_id = ?", user_id).Where("pod_id = ?", invite.PodID).First(&coll)
	if result.Error == nil {
		return AbstractError(c, "You already joined this wallet")
	}

	collaborator := Collaborator{
		UserID:     user_id,
		PodID:      invite.PodID,
		RoleTypeID: ROLE_TYPE_BASIC,
		Selector:   MakeSelector(COLLABORATOR_TABLE),
	}

	result = DB.Create(&collaborator)

	if result.Error != nil {
		return AbstractError(c, "Something went wrong")
	}

	// delete the invite, its been used
	DB.Delete(&invite)

	// send email to all collaborators
	SendWalletJoinedEmail(invite.PodID)

	return c.String(http.StatusOK, "Welcome to the party")
}

func GetJoinPod(c echo.Context) error {
	req := PodJoiner{}
	defer c.Request().Body.Close()
	err := json.NewDecoder(c.Request().Body).Decode(&req)
	if err != nil {
		return AbstractError(c, "Couldn't read request")
	}

	invite := Invite{}
	result := DB.Where("code = ?", req.Code).First(&invite)
	if result.Error != nil {
		return AbstractError(c, "This code is not valid")
	}

	log.Println("invite.PodID", invite.PodID)
	log.Println(invite.PodID)
	// get pod
	pod := Pod{}
	result = DB.Preload("Collaborators.User.Avatar").Preload("PayoutType").Preload("LifecycleType").First(&pod, invite.PodID)
	if result.Error != nil {
		return AbstractError(c, "Couldn't get wallet")
	}

	// return all avatars!
	var avatars [][]uint
	collaborators := pod.Collaborators

	for _, col := range collaborators {
		userAvatar := col.User.Avatar
		builtAvatar := AvatarToArray(userAvatar)
		avatars = append(avatars, builtAvatar)
	}

	podRes := JoiningPodAPI{
		Name:          pod.Name,
		Description:   pod.Description,
		Avatars:       avatars,
		PayoutType:    pod.PayoutType,
		LifecycleType: pod.LifecycleType,
	}

	return c.JSON(http.StatusOK, podRes)
}

func GetInvites(c echo.Context) error {
	// user_id, err := GetUserIdFromToken(c)
	// if err != nil {
	// 	return AbstractError(c,"Something went wrong")
	// }

	podSelector := c.Param("podSelector")

	pod := Pod{}
	result := DB.Where("selector = ?", podSelector).Find(&pod)
	if result.Error != nil {
		return AbstractError(c, "Something went wrong")
	}

	// get user
	invites := []InviteAPI{}
	result = DB.Model(&Invite{}).Where("pod_id = ?", pod.ID).Find(&invites)

	if result.Error != nil {
		return AbstractError(c, "Something went wrong")
	}

	return c.JSON(http.StatusOK, invites)
}

func DeleteInvite(c echo.Context) error {
	selector := c.Param("selector")

	// get user
	invite := Invite{}

	result := DB.Where("selector = ?", selector).First(&invite)
	if result.Error != nil {
		return AbstractError(c, "Something went wrong")
	}

	DB.Delete(&invite)

	return c.String(http.StatusOK, "OK!")
}

func ScheduleDestroyPod(c echo.Context) error {
	podSelector := c.Param("podSelector")

	pod := Pod{}
	result := DB.Where("selector = ?", podSelector).First(&pod)
	if result.Error != nil {
		return AbstractError(c, "Couldn't find user")
	}
	pod.ToDelete = time.Now().String()
	result = DB.Save(&pod)
	if result.Error != nil {
		return AbstractError(c, "Couldn't find user")
	}

	// send email to all collaborators
	SendWalletDestroyedEmail(podSelector)

	responseString := pod.Name + " will be deleted when payouts are finished."
	return c.String(http.StatusOK, responseString)
}
