package model

import (
	"database/sql/driver"

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
	Username    string
	Surname     string
	Email       string `gorm:"index:idx_user_email,unique"`
	DateOfBirth datatypes.Date
	IsVegan     bool
	Occupation  occupation `gorm:"type:occupation"`
	CreatedAt   datatypes.Time
}
type File struct {
	gorm.Model
	TeamID uint
	//TODO file storage
	CreatedAt datatypes.Time
	UpdatedAt datatypes.Time
}

// Enum occupation
type occupation string

const (
	UCZEN     occupation = "uczen"
	STUDENT   occupation = "student"
	ABSOLWENT occupation = "absolwent"
	INNE      occupation = "inne"
)

func (o *occupation) Scan(value any) error {
	*o = occupation(value.([]byte))
	return nil
}
func (o occupation) Value() (driver.Value, error) {
	return string(o), nil
}
