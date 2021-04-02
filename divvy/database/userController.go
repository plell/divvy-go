package database

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func parseRequestJSON(c echo.Context) (User, error) {
	decodedJson := User{}
	defer c.Request().Body.Close()
	err := json.NewDecoder(c.Request().Body).Decode(&decodedJson)
	if err != nil {
		log.Println("uuuuuuuu, $s", err)
		return decodedJson, c.String(http.StatusInternalServerError, "")
	}

	return decodedJson, nil
}

type UserCreator struct {
	DisplayName string `json:"displayName"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Feature1    uint   `json:"feature1"`
	Feature2    uint   `json:"feature2"`
	Feature3    uint   `json:"feature3"`
	Feature4    uint   `json:"feature4"`
	Feature5    uint   `json:"feature5"`
	Feature6    uint   `json:"feature6"`
	Feature7    uint   `json:"feature7"`
	Feature8    uint   `json:"feature8"`
	Feature9    uint   `json:"feature9"`
}

type CreateResponse struct {
	Token  string `json:"token"`
	User   User   `json:"user"`
	Avatar []uint `json:"avatar"`
}

type AvatarResponse struct {
	Avatar []uint `json:"avatar"`
}

func CreateUser(c echo.Context) error {
	req := UserCreator{}
	defer c.Request().Body.Close()
	err := json.NewDecoder(c.Request().Body).Decode(&req)

	if err != nil {
		return abstractError(c)
	}

	hashedPassword := hashAndSalt(req.Password)

	log.Println(hashedPassword)

	user := User{
		Username:    req.Username,
		Password:    hashedPassword,
		DisplayName: req.DisplayName,
	}

	result := DB.Create(&user) // pass pointer of data to Create

	if result.Error != nil {
		return abstractError(c)
	}

	avatar := Avatar{
		UserId:   user.ID,
		Feature1: req.Feature1,
		Feature2: req.Feature2,
		Feature3: req.Feature3,
		Feature4: req.Feature4,
		Feature5: req.Feature5,
		Feature6: req.Feature6,
		Feature7: req.Feature7,
		Feature8: req.Feature8,
		Feature9: req.Feature9}

	result = DB.Create(&avatar) // pass pointer of data to Create

	if result.Error != nil {
		return abstractError(c)
	}

	return c.String(http.StatusOK, "Success!")
}

func GetUser(c echo.Context) error {
	id := c.Param("userId")

	user := User{}

	result := DB.First(&user, id)

	if result.Error != nil {
		return abstractError(c)
	}

	return c.JSON(http.StatusOK, json.NewEncoder(c.Response()).Encode(user))
}

func UpdateAvatar(c echo.Context) error {
	req := Avatar{}
	defer c.Request().Body.Close()
	err := json.NewDecoder(c.Request().Body).Decode(&req)

	if err != nil {
		return abstractError(c)
	}

	avatar := Avatar{}
	// get avatar by user id
	result := DB.Where("user_id = ?", req.UserId).First(&avatar)
	if result.Error != nil {
		return abstractError(c)
	}
	// update
	result = DB.Model(&avatar).Updates(req)
	if result.Error != nil {
		return abstractError(c)
	}
	// get updated avatar
	result = DB.First(&avatar)
	if result.Error != nil {
		return abstractError(c)
	}

	avatarFeatures := []uint{avatar.Feature1,
		avatar.Feature2,
		avatar.Feature3,
		avatar.Feature4,
		avatar.Feature5,
		avatar.Feature6,
		avatar.Feature7,
		avatar.Feature8,
		avatar.Feature9}

	response := AvatarResponse{
		Avatar: avatarFeatures}

	return c.JSON(http.StatusOK, response)
}

// func DeleteUser(c echo.Context) error {
// 	// var user User
// 	// DB.db.First(&user, id)
// 	// DB.db.Delete(&user, id)
// }

// func UpdateUser(c echo.Context) error {
// 	// var user User
// 	// DB.db.First(&user, id)
// 	// DB.db.Model(&user).Update("Price", 200)
// 	// update multiple fields
// 	// db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})
// }

func hashAndSalt(pwd string) string {

	// Use GenerateFromPassword to hash & salt pwd
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)

	bytePwd := []byte(pwd)
	hash, err := bcrypt.GenerateFromPassword(bytePwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	// GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash)
}
