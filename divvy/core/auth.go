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

type GoogleCredentials struct {
	Email       string `json:"email"`
	GoogleID    string `json:"googleId"`
	AccessToken string `json:"accessToken"`
	TokenID     string `json:"tokenId"`
	City        string `json:"city"`
	Name        string `json:"name"`
	ImageURL    string `json:"imageUrl"`
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
	UserID       uint   `json:"userId"`
	UserSelector string `json:"userSelector"`
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
	outcome := "failed"
	if success {
		outcome = "succeeded"
	}
	LogInfo("Login " + outcome + " from " + ip)
}

// This is a basic login, not a google login
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
		return AbstractError(c, "Email or password incorrect")
	}

	// Check if password is correct
	if comparePasswords(user.Password, creds.Password) == false {
		// logged failed login
		MakeLoginHistory(creds.Username, ip, false)
		return AbstractError(c, "Email or password incorrect")
	}

	claims := &jwtCustomClaims{
		UserID:       user.ID,
		UserSelector: user.Selector,
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

	// login is correct! check if account is verified
	if user.Verified == "" {
		// if not, send verification email
		Direct_SendVerificationEmail(user)
	}

	return c.JSON(http.StatusOK, response)
}

// we know that the login was successful if they got here.
// step 1: check that access token is legit
// step 2: find user with this google id, if not found then create
// step 3: send back token and stuff
func GoogleLogin(c echo.Context) error {
	mySigningKey := GetSigningKey()
	ip := c.RealIP()

	// bind json to the login variable
	creds := GoogleCredentials{}
	defer c.Request().Body.Close()
	err := json.NewDecoder(c.Request().Body).Decode(&creds)
	if err != nil {
		log.Println("failed reading login request, $s", err)
		return c.String(http.StatusInternalServerError, "")
	}

	user := User{}

	if creds.GoogleID == "" || creds.Email == "" {
		return c.String(http.StatusInternalServerError, "Oops, that didn't work")
	}

	// check that token is legit
	tokenInfo, err := VerifyGoogleIdToken(creds.TokenID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Google verification failed")
	}

	// not only is the tokenId legit, it matches the user input
	if tokenInfo.Email != creds.Email {
		return c.String(http.StatusInternalServerError, "Google verification failed: mismatch")
	}

	// Check in your db if the user exists or not
	result := DB.Preload("Avatar").Where("username = ?", creds.Email).Where("google_id = ?", creds.GoogleID).First(&user)

	if result.Error != nil {
		// make a new user
		user_id, err := CreateGoogleUser(creds)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Email belongs to an existing account. Sign up manually or choose a different gmail account.")
		}

		result := DB.Preload("Avatar").First(&user, user_id)
		if result.Error != nil {
			return c.String(http.StatusInternalServerError, "Couldn't find newly created user")
		}

	}

	claims := &jwtCustomClaims{
		UserID:       user.ID,
		UserSelector: user.Selector,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * (24 * 7)).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	// The signing string should be secret (a generated UUID works too)
	t, err := token.SignedString(mySigningKey)
	if err != nil {
		return err
	}

	formatUser := BuildUser(user)

	response := LoginResponse{
		Token: t,
		User:  formatUser}

	MakeLoginHistory(creds.Email, ip, true)

	LogInfo("Google Login!")

	return c.JSON(http.StatusOK, response)
}

func CustomerLogin(c echo.Context) error {
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
	result := DB.Preload("Avatar").Preload("Customer").Where("username = ?", creds.Username).First(&user)

	if result.Error != nil {
		MakeLoginHistory(creds.Username, ip, false)
		return AbstractError(c, "Email or password incorrect")
	}

	// Check if password is correct
	if comparePasswords(user.Password, creds.Password) == false {
		// logged failed login
		MakeLoginHistory(creds.Username, ip, false)
		return AbstractError(c, "Email or password incorrect")
	}

	claims := &jwtCustomClaims{
		UserID:       user.ID,
		UserSelector: user.Selector,
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

type PasswordChangeReq struct {
	Code     string `json:"code"`
	Password string `json:"password"`
}

func ChangePassword(c echo.Context) error {
	passwordReq := PasswordChangeReq{}
	defer c.Request().Body.Close()
	err := json.NewDecoder(c.Request().Body).Decode(&passwordReq)
	if err != nil {
		log.Println("failed reading login request, $s", err)
		return c.String(http.StatusInternalServerError, "")
	}

	user := User{}
	result := DB.Where("password_reset_token = ?", passwordReq.Code).First(&user)
	if result.Error != nil {
		return AbstractError(c, "Token invalid.")
	}

	user.PasswordResetToken = ""
	user.PasswordLastChanged = time.Now().String()
	user.Password = HashAndSalt(passwordReq.Password)

	result = DB.Save(&user)
	if result.Error != nil {
		return AbstractError(c, "Couldn't save password.")
	}

	return c.String(http.StatusOK, "Success")
}

func Logout(c echo.Context) error {
	// user_id, err := GetUserIdFromToken(c)
	// if err != nil {
	// 	return AbstractError(c, "Something went wrong")
	// }

	return c.String(http.StatusOK, "Success")
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
	log.Println("GetUserIdFromToken")
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*jwtCustomClaims)
	log.Println(claims, "claims")
	user_id := claims.UserID

	log.Println("GOT USER ID ")
	log.Println(user_id)

	return user_id, nil
}
