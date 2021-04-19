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
	DisplayName   string        `json:"displayName"`
	Username      string        `json:"username"`
	City          string        `json:"city"`
	Password      string        `gorm:"varchar(70)" json:"password"`
	Selector      string        `json:"selector"`
	Verified      uint          `json:"verified"` // unix timestamp of when verified
	Avatar        Avatar        //`gorm:"PRELOAD"` //`gorm:"ForeignKey:ID;AssociationForeignKey:UserID"`
	StripeAccount StripeAccount //`gorm:"PRELOAD:false"`
	Collaborator  []Collaborator
	ByTheBy
	gorm.Model
}

type UserAPI struct {
	DisplayName string `json:"displayName"`
	Username    string `json:"username"`
	City        string `json:"city"`
	Selector    string `json:"selector"`
	Avatar      []uint `json:"avatar"`
}

var STRIPE_ACCOUT_TABLE = "stripe_accounts"

type StripeAccount struct {
	UserID   uint   `json:"userId"`
	AcctID   string `json:"acctId"`
	Selector string `json:"selector"`
	gorm.Model
	ByTheBy
}
type StripeAccountAPI struct {
	AcctID   string `json:"acctId"`
	Selector string `json:"selector"`
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
	User     User
	UserID   uint `json:"userId"`
	Pod      Pod
	PodID    uint    `json:"podId"`
	IsAdmin  bool    `json:"isAdmin"`
	Selector string  `json:"selector"`
	Claim    float64 `json:"claim"`
	gorm.Model
	ByTheBy
}
type CollaboratorAPI struct {
	IsAdmin     bool    `json:"isAdmin"`
	Selector    string  `json:"selector"`
	Claim       float64 `json:"claim"`
	DisplayName string  `json:"displayName"`
	Username    string  `json:"username"`
	City        string  `json:"city"`
	Avatar      []uint  `json:"avatar"`
}

var POD_TABLE = "pods"

type Pod struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	User        User
	UserID      uint   `json:"userId"`
	Selector    string `json:"selector"`
	gorm.Model
	ByTheBy
}
type PodAPI struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Selector    string `json:"selector"`
	MemberCount int    `json:"memberCount"`
}

var SELECTOR_TABLE = "selectors"

type Selector struct {
	Selector string `json:"selector"`
	Type     string `json:"type"`
	gorm.Model
	ByTheBy
}

var PAYMENT_TABLE = "payments"

type Payment struct {
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
	TransferGroup string `json:"transferGroup"`
	Status        uint   `json:"status"`
	SessionID     string `json:"sessionId"`
	Pod           Pod
	PodID         uint   `json:"PodId"`
	Selector      string `json:"selector"`
	gorm.Model
	ByTheBy
}
type PaymentAPI struct {
	ID       uint   `json:"id"`
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
	Selector string `json:"selector"`
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
