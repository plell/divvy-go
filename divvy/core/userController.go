package core

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type UserCreator struct {
	DisplayName string `json:"displayName"`
	BetaKey     string `json:"betaKey"`
	City        string `json:"city"`
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

	log.Println("CreateUser")
	if err != nil {
		return AbstractError(c, "Couldn't read request")
	}

	hashedPassword := HashAndSalt(req.Password)

	user := User{
		Username:    req.Username,
		Password:    hashedPassword,
		DisplayName: req.DisplayName,
		City:        req.City,
		Selector:    MakeSelector(USER_TABLE),
		UserTypeID:  USER_TYPE_BASIC,
	}

	result := DB.Create(&user) // pass pointer of data to Create

	if result.Error != nil {
		return AbstractError(c, result.Error.Error())
	}

	avatar := Avatar{
		UserID:   user.ID,
		Feature1: req.Feature1,
		Feature2: req.Feature2,
		Feature3: req.Feature3,
		Feature4: req.Feature4,
		Feature5: req.Feature5,
		Feature6: req.Feature6,
		Feature7: req.Feature7,
		Feature8: req.Feature8,
		Feature9: req.Feature9,
		Selector: MakeSelector(AVATAR_TABLE),
	}

	result = DB.Create(&avatar) // pass pointer of data to Create

	if result.Error != nil {
		return AbstractError(c, "Couldn't create avatar")
	}

	emailVerificationCode := EmailVerificationCode{
		UserID: user.ID,
		Code:   MakeInviteCode(),
	}

	result = DB.Create(&emailVerificationCode) // pass pointer of data to Create

	if result.Error != nil {
		return AbstractError(c, "Couldn't create verification code")
	}

	return c.String(http.StatusOK, "Success!")
}

func VerifyAccountEmail(c echo.Context) error {
	user_id, err := GetUserIdFromToken(c)
	if err != nil {
		return AbstractError(c, "Something went wrong")
	}

	code := c.Param("verificationCode")

	emailVC := EmailVerificationCode{}

	result := DB.Where("code = ?", code).First(&emailVC)
	if result.Error != nil {
		return AbstractError(c, "Verification code invalid")
	}

	if emailVC.UserID != user_id {
		return AbstractError(c, "Verification code invalid")
	}

	user := User{}

	result = DB.First(&user, user_id)
	if result.Error != nil {
		return AbstractError(c, "User not found")
	}

	DB.Model(&user).Update("verified", time.Now().String())

	return c.String(http.StatusOK, user.Verified)
}

func GetUser(c echo.Context) error {
	user_id, err := GetUserIdFromToken(c)
	if err != nil {
		return AbstractError(c, "Something went wrong")
	}

	user := User{}

	result := DB.Preload("Avatar").First(&user, user_id)

	if result.Error != nil {
		return AbstractError(c, "Something went wrong")
	}

	formatUser := BuildUser(user)

	return c.JSON(http.StatusOK, formatUser)
}

func GetAvatar(c echo.Context) error {
	user_id, err := GetUserIdFromToken(c)
	if err != nil {
		return AbstractError(c, "Something went wrong")
	}

	// find avatar
	avatar := Avatar{}
	result := DB.Where("user_id = ?", user_id).First(&avatar)
	if result.Error != nil {
		return AbstractError(c, "Something went wrong")
	}

	avatarFeatures := AvatarToArray(avatar)

	response := AvatarResponse{
		Avatar: avatarFeatures}

	return c.JSON(http.StatusOK, response)
}

func UpdateUser(c echo.Context) error {
	user_id, err := GetUserIdFromToken(c)
	if err != nil {
		return AbstractError(c, "Something went wrong")
	}

	req := User{}
	defer c.Request().Body.Close()
	err = json.NewDecoder(c.Request().Body).Decode(&req)

	// find user
	user := User{}
	result := DB.First(&user, user_id)
	if result.Error != nil {
		return AbstractError(c, "Couldn't find user")
	}
	// update display name
	user.DisplayName = req.DisplayName
	result = DB.Save(&user)
	if result.Error != nil {
		return AbstractError(c, "Couldn't save user")
	}

	return c.String(http.StatusOK, "Details saved!")
}

func UpdateAvatar(c echo.Context) error {
	user_id, err := GetUserIdFromToken(c)
	if err != nil {
		return AbstractError(c, "Something went wrong")
	}

	// decode request avatar details
	req := Avatar{}
	defer c.Request().Body.Close()
	err = json.NewDecoder(c.Request().Body).Decode(&req)

	if err != nil {
		return AbstractError(c, "Something went wrong")
	}

	avatar := Avatar{}
	// get avatar by user id
	result := DB.Where("user_id = ?", user_id).First(&avatar)
	if result.Error != nil {
		return AbstractError(c, "Something went wrong")
	}
	// update
	result = DB.Model(&avatar).Updates(req)
	if result.Error != nil {
		return AbstractError(c, "Something went wrong")
	}
	// get updated avatar
	result = DB.First(&avatar)
	if result.Error != nil {
		return AbstractError(c, "Something went wrong")
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

func HashAndSalt(pwd string) string {

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
