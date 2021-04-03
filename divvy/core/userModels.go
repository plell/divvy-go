package core

import (
	"gorm.io/gorm"
)

// gorm.Model injects id, deleted_at, created_at, and updated_at
type User struct {
	DisplayName string `json:"displayName"`
	Username    string `json:"username"`
	Password    string `gorm:"varchar(70)" json:"password"`
	gorm.Model
}

// this is the stripe connection needed to pay users
type StripeAccount struct {
	User   User
	UserId uint   `json:"userId"`
	AcctId string `json:"acctId"`
	gorm.Model
}

// stores avatar config for frontend
type Avatar struct {
	User     User
	UserId   uint `json:"userId"`
	Feature1 uint `json:"feature1"`
	Feature2 uint `json:"feature2"`
	Feature3 uint `json:"feature3"`
	Feature4 uint `json:"feature4"`
	Feature5 uint `json:"feature5"`
	Feature6 uint `json:"feature6"`
	Feature7 uint `json:"feature7"`
	Feature8 uint `json:"feature8"`
	Feature9 uint `json:"feature9"`
	gorm.Model
}

// a user has a collaborator record for each group theyre in
type Collaborator struct {
	User    User
	UserId  uint `json:"userId"`
	Group   Group
	GroupId uint `json:"groupId"`
	gorm.Model
}
