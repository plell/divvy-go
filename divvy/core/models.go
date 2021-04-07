package core

import (
	"gorm.io/gorm"
)

// gorm.Model injects id, deleted_at, created_at, and updated_at
var USER_TABLE = "users"

type User struct {
	DisplayName string `json:"displayName"`
	Username    string `json:"username"`
	City        string `json:"city"`
	Password    string `gorm:"varchar(70)" json:"password"`
	Selector    string `json:"selector"`
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
	User     User
	UserId   uint   `json:"userId"`
	AcctId   string `json:"acctId"`
	Selector string `json:"selector"`
	gorm.Model
}
type StripeAccountAPI struct {
	AcctId   string `json:"acctId"`
	Selector string `json:"selector"`
}

var AVATAR_TABLE = "avatars"

type Avatar struct {
	User     User
	UserId   uint   `json:"userId"`
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
	UserId   uint   `json:"userId"`
}

var COLLABORATOR_TABLE = "collaborators"

type Collaborator struct {
	User     User
	UserId   uint `json:"userId"`
	Pod      Pod
	PodId    uint    `json:"podId"`
	IsAdmin  uint    `json:"isAdmin"`
	Selector string  `json:"selector"`
	Claim    float64 `json:"claim"`
	gorm.Model
}
type CollaboratorAPI struct {
	ID       uint    `json:"id"`
	UserId   uint    `json:"userId"`
	PodId    uint    `json:"podId"`
	IsAdmin  uint    `json:"isAdmin"`
	Selector string  `json:"selector"`
	Claim    float64 `json:"claim"`
}

var POD_TABLE = "pods"

type Pod struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	User        User
	UserId      uint   `json:"userId"`
	Selector    string `json:"selector"`
	gorm.Model
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
}

var PAYMENT_TABLE = "payments"

type Payment struct {
	Amount         string `json:"amount"`
	Currency       string `json:"currency"`
	User           User
	UserId         uint `json:"userId"`
	Collaborator   Collaborator
	CollaboratorId uint `json:"collaboratorId"`
	Pod            Pod
	PodId          uint   `json:"podId"`
	Selector       string `json:"selector"`
	gorm.Model
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
	PodId       uint   `json:"podId"`
	CreatedById uint   `json:"createdById"`
	Selector    string `json:"selector"`
	gorm.Model
}

type InviteAPI struct {
	Email    string `json:"email"`
	Selector string `json:"selector"`
}
