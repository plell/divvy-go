package core

import (
	"gorm.io/gorm"
)

var USER_TABLE = "users"

// gorm.Model injects id, deleted_at, created_at, and updated_at
type User struct {
	DisplayName string `json:"displayName"`
	Username    string `json:"username"`
	Password    string `gorm:"varchar(70)" json:"password"`
	Selector    string `json:"selector"`
	gorm.Model
}

var STRIPE_ACCOUT_TABLE = "stripe_accounts"

// this is the stripe connection needed to pay users
type StripeAccount struct {
	User     User
	UserId   uint   `json:"userId"`
	AcctId   string `json:"acctId"`
	Selector string `json:"selector"`
	gorm.Model
}

var AVATAR_TABLE = "avatars"

// stores avatar config for frontend
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

var COLLABORATOR_TABLE = "collaborators"

// ties user to collab
type Collaborator struct {
	User     User
	UserId   uint `json:"userId"`
	Collab   Collab
	CollabId uint   `json:"collabId"`
	IsAdmin  uint   `json:"isAdmin"`
	Selector string `json:"selector"`
	gorm.Model
}

var COLLAB_TABLE = "collabs"

type Collab struct {
	Name        uint `json:"name"`
	Description uint `json:"description"`
	User        User
	UserId      uint   `json:"userId"`
	Selector    string `json:"selector"`
	gorm.Model
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
	Collab         Collab
	CollabId       uint   `json:"collabId"`
	Selector       string `json:"selector"`
	gorm.Model
}
