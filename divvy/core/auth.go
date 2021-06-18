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
	Username    string `json:"username"`
	GoogleID    string `json:"googleId"`
	AccessToken string `json:"accessToken"`
	TokenID     string `json:"tokenId"`
	City        string `json:"city"`
	Name        string `json:"name"`
	ImageURL    string `json:"imageUrl"`
	Password    string `json:"password"`
	BetaKey     string `json:"betaKey"`
	DisplayName string `json:"displayName"`
	Feature1    uint   `json:"feature1"`
	Feature2    uint   `json:"feature2"`
	Feature3    uint   `json:"feature3"`
	Feature4    uint   `json:"feature4"`
	Feature5    uint   `json:"feature5"`
	Feature6    uint   `json:"feature6"`
	Feature7    uint   `json:"feature7"`
	Feature8    uint   `json:"feature8"`
	Feature9    uint   `json:"feature9"`
	Feature10   uint   `json:"feature10"`
	Feature11   uint   `json:"feature11"`
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
	IsStore      bool   `json:"isStore"`
	IsApp        bool   `json:"isApp"`
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

	// make sure this is not a google id user
	if user.GoogleID != "" {
		// logged failed login
		MakeLoginHistory(creds.Username, ip, false)
		return AbstractError(c, "Please use Google Sign In")
	}

	if user.BetaKey == "" {
		errorstring, errbool := DoBetaKeyCheck(creds)
		if errbool {
			return c.String(http.StatusUnauthorized, "Beta key expired")
		}
		if errorstring != "" {
			return c.String(http.StatusOK, errorstring)
		}

		// add beta key to user, also verified (because they got the beta key by email)
		user.BetaKey = creds.BetaKey
		user.Verified = time.Now().String()
		result = DB.Save(&user)
		if result.Error != nil {
			return c.String(http.StatusUnauthorized, "Couldn't update user")
		}
		DeleteBetaKey(creds)
	}

	claims := &jwtCustomClaims{
		UserID:       user.ID,
		UserSelector: user.Selector,
		IsApp:        true,
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
func GoogleLoginOrSignUp(c echo.Context) error {
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

	if creds.GoogleID == "" || creds.Username == "" {
		return c.String(http.StatusInternalServerError, "Oops, that didn't work")
	}

	// check that token is legit
	tokenInfo, err := VerifyGoogleIdToken(creds.TokenID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Google verification failed")
	}

	// not only is the tokenId legit, it matches the user input
	if tokenInfo.Email != creds.Username {
		return c.String(http.StatusInternalServerError, "Google verification failed: mismatch")
	}

	// Check in your db if the user exists or not
	result := DB.Preload("Avatar").Where("username = ?", creds.Username).Where("google_id = ?", creds.GoogleID).First(&user)

	if result.Error != nil {
		// make a new user
		// ************* TEMPORARY BETA REQUIREMENT START
		errorstring, errbool := DoBetaKeyCheck(creds)
		if errbool {
			return c.String(http.StatusUnauthorized, "Beta key expired")
		}
		if errorstring != "" {
			return c.String(http.StatusOK, errorstring)
		}
		// ************* TEMPORARY BETA REQUIREMENT END

		user_id, err := CreateGoogleUser(creds)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Email belongs to an existing account. Sign up manually or choose a different gmail account.")
		}

		result := DB.Preload("Avatar").First(&user, user_id)
		if result.Error != nil {
			return c.String(http.StatusInternalServerError, "Couldn't find newly created user")
		}
	} else {
		// if user was created as a customer, now using the app, we need to require a betakey
		// ************* TEMPORARY BETA REQUIREMENT START
		if user.BetaKey == "" {
			errorstring, errbool := DoBetaKeyCheck(creds)
			if errbool {
				return c.String(http.StatusUnauthorized, "Beta key expired")
			}
			if errorstring != "" {
				return c.String(http.StatusOK, errorstring)
			}

			// add beta key to user
			user.BetaKey = creds.BetaKey
			DB.Save(&user)
		}
		// ************* TEMPORARY BETA REQUIREMENT END
	}

	claims := &jwtCustomClaims{
		UserID:       user.ID,
		UserSelector: user.Selector,
		IsApp:        true,
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

	MakeLoginHistory(creds.Username, ip, true)

	DeleteBetaKey(creds)

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

	// make sure this is not a google id user
	if user.GoogleID != "" {
		// logged failed login
		MakeLoginHistory(creds.Username, ip, false)
		return AbstractError(c, "Please use Google Sign In")
	}

	claims := &jwtCustomClaims{
		UserID:       user.ID,
		UserSelector: user.Selector,
		IsStore:      true,
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

func CustomerGoogleLoginOrSignUp(c echo.Context) error {
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

	if creds.GoogleID == "" || creds.Username == "" {
		return c.String(http.StatusInternalServerError, "Oops, that didn't work")
	}

	// check that token is legit
	tokenInfo, err := VerifyGoogleIdToken(creds.TokenID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Google verification failed")
	}

	// not only is the tokenId legit, it matches the user input
	if tokenInfo.Email != creds.Username {
		return c.String(http.StatusInternalServerError, "Google verification failed: mismatch")
	}

	// Check in your db if the user exists or not
	result := DB.Preload("Avatar").Where("username = ?", creds.Username).Where("google_id = ?", creds.GoogleID).First(&user)

	if result.Error != nil {
		// make a new user
		user_id, err := CreateGoogleUser(creds)
		if err != nil {
			return c.String(http.StatusInternalServerError, "Email belongs to an existing account. Sign up manually or choose a different gmail account.")
		}

		result := DB.First(&user, user_id)
		if result.Error != nil {
			return c.String(http.StatusInternalServerError, "Couldn't find newly created user")
		}
	}

	claims := &jwtCustomClaims{
		UserID:       user.ID,
		IsStore:      true,
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

	MakeLoginHistory(creds.Username, ip, true)

	LogInfo("Google Customer Login!")

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

	// we dont reset google users
	if user.GoogleID != "" {
		return AbstractError(c, "This account's password is managed by Google - it can't be reset here.")
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

func DoBetaKeyCheck(req Credentials) (string, bool) {
	// at this point, we know that the user does not exist
	// we are creating a new user
	// if they have a beta key, validate and pass
	if req.BetaKey == "" {
		return "betaKeyRequired", false
	}

	// check that beta key exists
	betaKey := BetaKey{}
	result := DB.Where("beta_key = ?", req.BetaKey).First(&betaKey)
	if result.Error != nil {
		// if no beta key exists, check invite codes
		inviteCode := Invite{}
		result := DB.Where("code = ?", req.BetaKey).First(&inviteCode)
		if result.Error != nil {
			return "", true
		}
	}

	return "", false
}

func DeleteBetaKey(req Credentials) {
	// check that beta key exists
	if req.BetaKey != "" {
		betaKey := BetaKey{}
		result := DB.Where("beta_key = ?", req.BetaKey).First(&betaKey)
		if result.Error == nil {
			DB.Delete(&betaKey)
		}
	}
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
