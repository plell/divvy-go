package database

import (
	"gorm.io/gorm"
)

// gorm.Model injects id, deleted_at, created_at, and updated_at
type Group struct {
	User     User
	UserId   uint   `json:"userId"`
	Currency string `json:"currency"`
	gorm.Model
}

type Payment struct {
	Amount         string `json:"amount"`
	Currency       string `json:"currency"`
	User           User
	UserId         uint `json:"userId"`
	Collaborator   Collaborator
	CollaboratorId uint `json:"collaboratorId"`
	Group          Group
	GroupId        uint `json:"groupId"`
	gorm.Model
}
