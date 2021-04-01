package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	db "github.com/plell/divvygo/divvy/database"
	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token  string  `json:"token"`
	User   db.User `json:"user"`
	Avatar []uint  `json:"avatar"`
}

// Most of the code is taken from the echo guide
// https://echo.labstack.com/cookbook/jwt
func Login(c echo.Context) error {

	// bind json to the login variable
	creds := Credentials{}
	defer c.Request().Body.Close()
	err := json.NewDecoder(c.Request().Body).Decode(&creds)
	if err != nil {
		log.Println("failed reading login request, $s", err)
		return c.String(http.StatusInternalServerError, "")
	}

	user := db.User{}

	// Check in your db if the user exists or not
	result := db.DB.Where("username = ?", creds.Username).First(&user)

	if result.Error != nil {
		return echo.ErrUnauthorized
	}

	// Check if password is correct
	if comparePasswords(user.Password, creds.Password) == false {
		return echo.ErrUnauthorized
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)
	// Set claims
	// This is the information which frontend can use
	// The backend can also decode the token and get admin etc.
	claims := token.Claims.(jwt.MapClaims)
	claims["name"] = "David Plell"
	claims["admin"] = true
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	// Generate encoded token and send it as response.
	// The signing string should be secret (a generated UUID          works too)
	t, err := token.SignedString(mySigningKey)
	if err != nil {
		return err
	}

	// Check in your db if the user exists or not
	avatar := db.Avatar{}

	result = db.DB.Where("user_id = ?", user.ID).First(&avatar)
	if result.Error != nil {
		return echo.ErrUnauthorized
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

	response := LoginResponse{
		Token:  t,
		User:   user,
		Avatar: avatarFeatures}

	return c.JSON(http.StatusOK, response)

}

func comparePasswords(hashedPwd string, plainPwd string) bool {
	// Since we'll be getting the hashed password from the DB it
	// will be a string so we'll need to convert it to a byte slice
	log.Println("comparePasswords")

	byteHash := []byte(hashedPwd)
	bytePlain := []byte(plainPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, bytePlain)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}
