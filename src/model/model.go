package model

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Team struct {
	gorm.Model
	TeamName  string `gorm:"uniqueIndex"`
	Password  string
	CreatedAt datatypes.Time
	UpdatedAt datatypes.Time
	Users     []User
	File      File
}

type User struct {
	gorm.Model
	TeamID      uint
	Username    string         `json:"firstName" binding:"required"`
	Surname     string         `json:"lastName" binding:"required"`
	Email       string         `gorm:"index:idx_user_email,unique,serializer:json"`
	DateOfBirth datatypes.Date `json:"dateOfBirth" binding:"required"`
	IsVegan     bool           `json:"isVegan" binding:"required"`
	Agreement   bool           `json:"agreement" binding:"required"`
	Occupation  string         `json:"occupation" binding:"required"`
	CreatedAt   datatypes.Time
}
type File struct {
	gorm.Model
	TeamID uint
	//TODO file storage
	CreatedAt datatypes.Time
	UpdatedAt datatypes.Time
}
