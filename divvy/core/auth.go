package core

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string  `json:"token"`
	User  UserAPI `json:"user"`
}

type jwtUserClaims struct {
	User
	jwt.StandardClaims
}

type jwtCustomClaims struct {
	UserID uint `json:"UserId"`
	// UUID  string `json:"uuid"`
	// Admin bool   `json:"admin"`
	jwt.StandardClaims
}

func GetSigningKey() []byte {
	mySigningKey := []byte(os.Getenv("JWT_TOKEN"))
	return mySigningKey
}

func MakeLoginHistory(username string, ip string, success bool) {
	lh := LoginHistory{
		IP:       ip,
		Username: username,
		Success:  success,
	}

	DB.Create(&lh)
}

// Most of the code is taken from the echo guide
// https://echo.labstack.com/cookbook/jwt
func Login(c echo.Context) error {
	mySigningKey := GetSigningKey()
	ip := c.RealIP()

	// bind json to the login variable
	creds := Credentials{}
	defer c.Request().Body.Close()
	err := json.NewDecoder(c.Request().Body).Decode(&creds)
	if err != nil {
		log.Println("failed reading login request, $s", err)
		return c.String(http.StatusInternalServerError, "")
	}

	user := User{}

	// Check in your db if the user exists or not
	result := DB.Preload("Avatar").Where("username = ?", creds.Username).First(&user)

	if result.Error != nil {
		MakeLoginHistory(creds.Username, ip, false)
		return echo.ErrUnauthorized
	}

	// Check if password is correct
	if comparePasswords(user.Password, creds.Password) == false {
		// logged failed login
		MakeLoginHistory(creds.Username, ip, false)
		return echo.ErrUnauthorized
	}
	// login is correct! check if account is verified
	if user.Verified == 0 {
		// if not, send verification email
		SendVerificationEmail(user)
	}

	claims := &jwtCustomClaims{
		UserID: user.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * (24 * 7)).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	// The signing string should be secret (a generated UUID          works too)
	t, err := token.SignedString(mySigningKey)
	if err != nil {
		return err
	}

	formatUser := BuildUser(user)

	response := LoginResponse{
		Token: t,
		User:  formatUser}

	MakeLoginHistory(creds.Username, ip, true)

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

func GetUserIdFromToken(c echo.Context) (uint, error) {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*jwtCustomClaims)
	log.Println(claims, "claims")
	user_id := claims.UserID

	log.Println("GOT USER ID ")
	log.Println(user_id)

	return user_id, nil
}
