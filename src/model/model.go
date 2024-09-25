package model

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Team struct {
	gorm.Model
	TeamName          string `gorm:"uniqueIndex"`
	VerificationToken string
	Members           []Member
}

type Member struct {
	gorm.Model
	TeamID         uint
	Email          string `gorm:"index:idx_user_email,unique"`
	Password       string
	FirstName      *string
	LastName       *string
	DateOfBirth    *datatypes.Date
	Occupation     *string
	DietPrefernces *string
	Aggrement      bool
	School         *string
	IsVerified     bool
}
