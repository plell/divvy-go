package core

import (
	"gorm.io/gorm"
)

type ByTheBy struct {
	DeletedByID uint `json:"deleted_by_id"`
	CreatedByID uint `json:"created_by_id"`
	UpdatedByID uint `json:"updated_by_id"`
}

// gorm.Model injects id, deleted_at, created_at, and updated_at
var LOGIN_HISTORY_TABLE = "login_histories"

type LoginHistory struct {
	Username string `json:"username"`
	IP       string `json:"ip"`
	Success  bool   `json:"success"`
	ByTheBy
	gorm.Model
}

var USER_TABLE = "users"

type User struct {
	DisplayName           string `json:"displayName"`
	Username              string `gorm:"type:varchar(100);unique_index;unique" json:"username"`
	City                  string `json:"city"`
	Password              string `json:"password"`
	PasswordResetToken    string `json:"passwordResetToken"`
	PasswordLastChanged   string `json:"passwordLastChanged"`
	UserType              UserType
	UserTypeID            uint          `json:"userTypeId"`
	Selector              string        `json:"selector"`
	Verified              string        `json:"verified"` // datestring of when verified
	Avatar                Avatar        //`gorm:"PRELOAD"`  //`gorm:"ForeignKey:ID;AssociationForeignKey:UserID"`
	StripeAccount         StripeAccount //`gorm:"PRELOAD:false"`
	Collaborator          []Collaborator
	EmailVerificationCode EmailVerificationCode
	ByTheBy
	gorm.Model
}

type UserAPI struct {
	DisplayName string `json:"displayName"`
	Username    string `json:"username"`
	UserTypeID  uint   `json:"userTypeId"`
	City        string `json:"city"`
	Verified    string `json:"verified"` // datestring of when verified
	Selector    string `json:"selector"`
	Avatar      []uint `json:"avatar"`
}

var STRIPE_ACCOUT_TABLE = "stripe_accounts"

type StripeAccount struct {
	UserID   uint   `json:"userId"`
	AcctID   string `json:"acctId"`
	Selector string `json:"selector"`
	Verified string `json:"verified"`
	gorm.Model
	ByTheBy
}
type StripeAccountAPI struct {
	AcctID   string `json:"acctId"`
	Selector string `json:"selector"`
	Verified string `json:"verified"`
}

var AVATAR_TABLE = "avatars"

type Avatar struct {
	UserID   uint   `json:"userId"`
	Feature1 uint   `json:"feature1"`
	Feature2 uint   `json:"feature2"`
	Feature3 uint   `json:"feature3"`
	Feature4 uint   `json:"feature4"`
	Feature5 uint   `json:"feature5"`
	Feature6 uint   `json:"feature6"`
	Feature7 uint   `json:"feature7"`
	Feature8 uint   `json:"feature8"`
	Feature9 uint   `json:"feature9"`
	Selector string `json:"selector"`
	gorm.Model
	ByTheBy
}
type AvatarAPI struct {
	Feature1 uint   `json:"feature1"`
	Feature2 uint   `json:"feature2"`
	Feature3 uint   `json:"feature3"`
	Feature4 uint   `json:"feature4"`
	Feature5 uint   `json:"feature5"`
	Feature6 uint   `json:"feature6"`
	Feature7 uint   `json:"feature7"`
	Feature8 uint   `json:"feature8"`
	Feature9 uint   `json:"feature9"`
	Selector string `json:"selector"`
	UserID   uint   `json:"userId"`
}

var COLLABORATOR_TABLE = "collaborators"

type Collaborator struct {
	User         User    //`gorm:"PRELOAD:true"`
	UserID       uint    `json:"userId"`
	PodID        uint    `json:"podId"`
	Selector     string  `json:"selector"`
	Distribution float64 `json:"distribution"`
	RoleType     RoleType
	RoleTypeID   uint `json:"roleTypeId"`
	gorm.Model
	ByTheBy
}
type CollaboratorAPI struct {
	IsAdmin      bool    `json:"isAdmin"`
	Selector     string  `json:"selector"`
	UserSelector string  `json:"userSelector"`
	DisplayName  string  `json:"displayName"`
	Username     string  `json:"username"`
	Distribution float64 `json:"distribution"`
	City         string  `json:"city"`
	Avatar       []uint  `json:"avatar"`
	RoleType     RoleType
	RoleTypeID   uint `json:"roleTypeId"`
}

var POD_TABLE = "pods"

type Pod struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	User            User
	UserID          uint             `json:"userId"`
	Selector        string           `json:"selector"`
	PayoutType      PodPayoutType    `json:"payoutType"`
	PayoutTypeId    uint             `json:"payoutTypeId"`
	LifecycleType   PodLifecycleType `json:"lifecycleType"`
	LifecycleTypeId uint             `json:"lifecycleTypeId"`
	ToDelete        string           `json:"toDelete"`
	Collaborators   []Collaborator
	PodRule         []PodRule
	gorm.Model
	ByTheBy
}
type PodAPI struct {
	ID              uint             `json:"id"`
	Name            string           `json:"name"`
	Description     string           `json:"description"`
	Selector        string           `json:"selector"`
	ToDelete        string           `json:"toDelete"`
	PayoutTypeId    uint             `json:"payoutTypeId"`
	LifecycleTypeId uint             `json:"lifecycleTypeId"`
	MemberCount     int              `json:"memberCount"`
	PayoutType      PodPayoutType    `json:"payoutType"`
	LifecycleType   PodLifecycleType `json:"lifecycleType"`
}

type JoiningPodAPI struct {
	Name          string           `json:"name"`
	Description   string           `json:"description"`
	Avatars       [][]uint         `json:"avatars"`
	MemberCount   int              `json:"memberCount"`
	PayoutType    PodPayoutType    `json:"payoutType"`
	LifecycleType PodLifecycleType `json:"lifecycleType"`
}

// add pod rules relational table
var POD_RULE_TABLE = "pod_rules"

type PodRule struct {
	Value         string `json:"value"`
	PodRuleTypeID uint   `json:"podRuleTypeId"`
	PodID         uint   `json:"podID"`
	gorm.Model
	ByTheBy
}

var SELECTOR_TABLE = "selectors"

type Selector struct {
	Selector string `json:"selector"`
	Type     string `json:"type"`
	gorm.Model
	ByTheBy
}

var INVITE_TABLE = "invites"

type Invite struct {
	Code        string `json:"code"`
	Email       string `json:"email"`
	Pod         Pod
	PodID       uint   `json:"podId"`
	CreatedByID uint   `json:"createdById"`
	Selector    string `json:"selector"`
	gorm.Model
	ByTheBy
}

type InviteAPI struct {
	Email    string `json:"email"`
	Selector string `json:"selector"`
}

var EMAIL_VERIFICATION_CODE_TABLE = "email_verification_codes"

type EmailVerificationCode struct {
	Code   string `json:"code"`
	UserID uint   `json:"userId"`
	gorm.Model
	ByTheBy
}

// ***************static tables!

// add pod rules "maxPrice", "minPrice"
var RULE_TYPE_TABLE = "pod_rule_types"

type PodRuleType struct {
	Name string `json:"name"`
	ID   uint   `json:"id"`
}

// add pod rules "maxPrice", "minPrice"
var USER_TYPE_TABLE = "user_types"

type UserType struct {
	Name string `json:"name"`
	ID   uint   `json:"id"`
}

// add payout types "even", "admin25", "admin50", "admin75",
var POD_PAYOUT_TYPE_TABLE = "pod_payout_types"

type PodPayoutType struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ID          uint   `json:"id"`
}

// add pod rules "maxPrice", "minPrice"
var POD_LIFECYCLE_TYPE_TABLE = "pod_lifecycle_types"

type PodLifecycleType struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ID          uint   `json:"id"`
}

var ROLE_TYPE_TABLE = "role_types"

type RoleType struct {
	Name string `json:"name"`
	ID   uint   `json:"id"`
}

var BETA_KEY_TABLE = "beta_keys"

type BetaKey struct {
	BetaKey string `json:"betaKey"`
	gorm.Model
	ByTheBy
}

var BETA_KEY_REQUESTS_TABLE = "beta_key_requests"

type BetaKeyRequest struct {
	AcceptCode string `json:"acceptCode"`
	Email      string `json:"email"`
	Message    string `json:"message"`
	gorm.Model
	ByTheBy
}
