package model

import (
	"database/sql/driver"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Team struct {
	gorm.Model
	TeamName  string         // A regular string field
	Password  string         // A pointer to a string, allowing for null values
	CreatedAt datatypes.Time // Automatically managed by GORM for creation time
	UpdatedAt datatypes.Time // Automatically managed by GORM for update time
	Users     []User
	File      File
}

type User struct {
	gorm.Model
	TeamID      uint
	Username    string
	Surname     string
	Email       string
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
