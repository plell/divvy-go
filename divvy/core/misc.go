package core

import (
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/leekchan/accounting"
	_ "github.com/leekchan/accounting"
	_ "github.com/shopspring/decimal"
)

func Pong(c echo.Context) error {
	return c.String(http.StatusOK, "Pong")
}

func AbstractError(c echo.Context, message string) error {
	return c.String(http.StatusInternalServerError, message)
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
		avatar.Feature9,
		avatar.Feature10,
		avatar.Feature11}

	return avatarFeatures
}

func BuildUser(user User) UserAPI {
	return UserAPI{
		DisplayName: user.DisplayName,
		Username:    user.Username,
		UserTypeID:  user.UserTypeID,
		Selector:    user.Selector,
		City:        user.City,
		Verified:    user.Verified,
		Avatar:      AvatarToArray(user.Avatar),
	}
}

// build user from collaborator
func BuildUserFromCollaborator(collaborator Collaborator) CollaboratorAPI {
	u := collaborator.User
	//fixme, check
	hasStripeAccount := true
	if u.StripeAccount.AcctID == "" {
		hasStripeAccount = false
	}

	return CollaboratorAPI{
		DisplayName:      u.DisplayName,
		Username:         u.Username,
		Selector:         collaborator.Selector,
		UserSelector:     u.Selector,
		City:             u.City,
		HasStripeAccount: hasStripeAccount,
		IsAdmin:          collaborator.RoleTypeID == ROLE_TYPE_ADMIN,
		Avatar:           AvatarToArray(u.Avatar),
		RoleTypeID:       collaborator.RoleTypeID,
	}
}

func BuildPod(pod Pod) PodAPI {
	collaborators := pod.Collaborators
	memberCount := len(collaborators)
	var memberAvatars [][]uint

	for _, c := range collaborators {
		av := AvatarToArray(c.User.Avatar)
		memberAvatars = append(memberAvatars, av)
	}

	// add total earned and pending from stripe...
	totalPending, totalEarned := Direct_GetPodEarningsAndPendingTotal(pod.Selector)

	// totalPending := Direct_GetPodPendingTotal(pod.Selector)

	return PodAPI{
		Name:          pod.Name,
		Description:   pod.Description,
		Selector:      pod.Selector,
		MemberCount:   memberCount,
		MemberAvatars: memberAvatars,
		PayoutType:    pod.PayoutType,
		LifecycleType: pod.LifecycleType,
		ToDelete:      pod.ToDelete,
		TotalEarned:   totalEarned,
		TotalPending:  totalPending,
	}
}

func FormatAmountToString(amount int64, symbol string) string {
	// p := strconv.Itoa(int(amount))

	af := float64(amount) / 100

	ac := accounting.Accounting{Symbol: symbol, Precision: 2}

	a := ac.FormatMoney(af)

	return a
}

func FormatStringAmountNoSymbol(amount string) string {
	p, _ := strconv.Atoi(amount)

	af := float64(p) / 100

	ac := accounting.Accounting{Symbol: "", Precision: 2}

	a := ac.FormatMoney(af)

	return a
}

type AvatarOptions struct {
	TopType         []string `json:"topType"`
	AccessoriesType []string `json:"accessoryType"`
	HairColor       []string `json:"hairColor"`
	FacialHairType  []string `json:"facialHairType"`
	ClotheType      []string `json:"clotheType"`
	EyeType         []string `json:"eyeType"`
	EyebrowType     []string `json:"eyebrowType"`
	MouthType       []string `json:"mouthType"`
	SkinColor       []string `json:"skinColor"`
	FacialHairColor []string `json:"facialHairColor"`
	ClotheColor     []string `json:"clotheColor"`
}

func GetAvatarOptions(c echo.Context) error {
	avatarOptions := AvatarOptions{}
	avatarOptions.TopType = []string{
		"NoHair",
		"Eyepatch",
		"Hat",
		"Hijab",
		"Turban",
		"WinterHat1",
		"WinterHat2",
		"WinterHat3",
		"WinterHat4",
		"LongHairBigHair",
		"LongHairBob",
		"LongHairBun",
		"LongHairCurly",
		"LongHairCurvy",
		"LongHairDreads",
		"LongHairFrida",
		"LongHairFro",
		"LongHairFroBand",
		"LongHairNotTooLong",
		"LongHairShavedSides",
		"LongHairMiaWallace",
		"LongHairStraight",
		"LongHairStraight2",
		"LongHairStraightStrand",
		"ShortHairDreads01",
		"ShortHairDreads02",
		"ShortHairFrizzle",
		"ShortHairShaggyMullet",
		"ShortHairShortCurly",
		"ShortHairShortFlat",
		"ShortHairShortRound",
		"ShortHairShortWaved",
		"ShortHairSides",
		"ShortHairTheCaesar",
		"ShortHairTheCaesarSidePart",
	}
	avatarOptions.AccessoriesType = []string{
		"Blank",
		"Kurt",
		"Prescription01",
		"Prescription02",
		"Round",
		"Sunglasses",
		"Wayfarers",
	}
	avatarOptions.HairColor = []string{
		"Auburn",
		"Black",
		"Blonde",
		"BlondeGolden",
		"Brown",
		"BrownDark",
		"PastelPink",
		"Platinum",
		"Red",
		"SilverGray",
	}
	avatarOptions.FacialHairType = []string{
		"Blank",
		"BeardMedium",
		"BeardLight",
		"BeardMagestic",
		"MoustacheFancy",
		"MoustacheMagnum",
	}
	avatarOptions.ClotheType = []string{
		"BlazerShirt",
		"BlazerSweater",
		"CollarSweater",
		"GraphicShirt",
		"Hoodie",
		"Overall",
		"ShirtCrewNeck",
		"ShirtScoopNeck",
		"ShirtVNeck",
	}
	avatarOptions.EyeType = []string{
		"Close",
		"Default",
		"Dizzy",
		"EyeRoll",
		"Happy",
		"Hearts",
		"Side",
		"Squint",
		"Surprised",
		"Wink",
		"WinkWacky",
	}
	avatarOptions.EyebrowType = []string{
		"Angry",
		"AngryNatural",
		"Default",
		"DefaultNatural",
		"FlatNatural",
		"RaisedExcited",
		"RaisedExcitedNatural",
		"SadConcerned",
		"SadConcernedNatural",
		"UnibrowNatural",
		"UpDown",
		"UpDownNatural",
	}
	avatarOptions.MouthType = []string{
		"Concerned",
		"Default",
		"Disbelief",
		"Eating",
		"Grimace",
		"Sad",
		"ScreamOpen",
		"Serious",
		"Smile",
		"Tongue",
		"Twinkle",
	}
	avatarOptions.SkinColor = []string{
		"Brown",
		"Tanned",
		"Yellow",
		"Black",
		"Pale",
		"Light",
		"DarkBrown",
	}

	avatarOptions.ClotheColor = []string{
		"Black",
		"Blue01",
		"Blue02",
		"Blue03",
		"Gray01",
		"Gray02",
		"Heather",
		"PastelBlue",
		"PastelGreen",
		"PastelRed",
		"PastelYellow",
		"PastelOrange",
		"Pink",
		"White",
		"Red",
	}

	avatarOptions.FacialHairColor = []string{
		"Auburn",
		"Black",
		"Blonde",
		"BlondeGolden",
		"Brown",
		"BrownDark",
		"Platinum",
		"Red",
	}

	return c.JSON(http.StatusOK, avatarOptions)
}
