package model

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// User and Team processing DB Obects
type Team struct {
	gorm.Model
	TeamName          string `gorm:"uniqueIndex"`
	VerificationToken string
	IsVerified        bool
	IsConfirmed       bool     `gorm:"default:false"`
	ApproveStatus     string   `gorm:"default:pending"`
	Members           []Member `gorm:"foreignKey:team_id"`
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
	Agreement      bool
	School         *string
	IsVerified     bool
}

// Admin account DB Object
type HackArenaAdmin struct {
	gorm.Model
	Name      string
	Email     string `gorm:"index:idx_admin_email,unique"`
	User      string `gorm:"index:idx_admin_name,unique"`
	Password  string
	Privilage string `gorm:"default:SuperUser"`
}
