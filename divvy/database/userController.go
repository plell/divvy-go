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

	// result := DB.db.Create(&user) // pass pointer of data to Create
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

	DB.Create(&avatar) // pass pointer of data to Create

	// user.ID             // returns inserted data's primary key
	// result.Error        // returns error
	// result.RowsAffected // returns inserted records count

	return c.JSON(http.StatusOK, json.NewEncoder(c.Response()).Encode(user))
}

func GetUser(c echo.Context) error {
	id := c.Param("userId")

	log.Println("id")
	log.Println(id)

	user := User{}

	result := DB.First(&user, id)

	if result.Error != nil {
		return abstractError(c)
	}

	return c.JSON(http.StatusOK, json.NewEncoder(c.Response()).Encode(user))
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
