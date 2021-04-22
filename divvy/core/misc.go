package core

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func Pong(c echo.Context) error {
	return c.String(http.StatusOK, "Pong")
}

func AbstractError(c echo.Context) error {
	return c.String(http.StatusInternalServerError, "")
}

var pool = "abcdefghijklmnopqrstuvwxyzABCEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func MakeSelector(tableName string) string {
	rand.Seed(time.Now().UnixNano())
	l := 24
	bytes := make([]byte, l)

	randomSelector := ""
	// enter while loop, exited when n = 2
	n := 0
	for n < 1 {
		// create random string
		for i := 0; i < l; i++ {
			bytes[i] = pool[rand.Intn(len(pool))]
		}

		randomSelector = string(bytes)
		selector := Selector{}

		// create record in selectors to make sure only unique selector are made
		result := DB.Table(tableName).Where("selector = ?", randomSelector).First(&selector)
		if result.Error != nil {
			// good, this is a unique selector
			selector := Selector{
				Selector: randomSelector,
				Type:     tableName,
			}
			result := DB.Create(&selector) // pass pointer of data to Create
			if result.Error != nil {
				// db create failed
			}
			// leave loop
			log.Println("Made unique selector")
			n++
		} else {
			log.Println("Made duplicate selector, retry")
		}
	}

	return randomSelector
}

func MakeInviteCode() string {
	rand.Seed(time.Now().UnixNano())
	l := 24
	bytes := make([]byte, l)

	randomSelector := ""
	// create random string
	for i := 0; i < l; i++ {
		bytes[i] = pool[rand.Intn(len(pool))]
	}

	randomSelector = string(bytes)

	return randomSelector
}

func ContainsInt(arr []uint, val uint) bool {
	for _, a := range arr {
		if a == val {
			return true
		}
	}
	return false
}

func AvatarToArray(avatar Avatar) []uint {
	avatarFeatures := []uint{avatar.Feature1,
		avatar.Feature2,
		avatar.Feature3,
		avatar.Feature4,
		avatar.Feature5,
		avatar.Feature6,
		avatar.Feature7,
		avatar.Feature8,
		avatar.Feature9}

	return avatarFeatures
}

func BuildUser(user User) UserAPI {
	return UserAPI{
		DisplayName: user.DisplayName,
		Username:    user.Username,
		Selector:    user.Selector,
		City:        user.City,
		Avatar:      AvatarToArray(user.Avatar),
	}
}

// build user from collaborator
func BuildUserFromCollaborator(collaborator Collaborator) CollaboratorAPI {
	u := collaborator.User
	return CollaboratorAPI{
		DisplayName:  u.DisplayName,
		Username:     u.Username,
		Selector:     collaborator.Selector,
		UserSelector: u.Selector,
		City:         u.City,
		IsAdmin:      collaborator.IsAdmin,
		Avatar:       AvatarToArray(u.Avatar),
	}
}

func BuildPod(pod Pod) PodAPI {
	collaborators := []Collaborator{}
	DB.Model(&Collaborator{}).Where("pod_id = ?", pod.ID).Find(&collaborators)

	memberCount := len(collaborators)
	return PodAPI{
		Name:        pod.Name,
		Description: pod.Description,
		Selector:    pod.Selector,
		MemberCount: memberCount,
	}
}
