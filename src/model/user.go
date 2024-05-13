package model

import (
	"database/sql/driver"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	TeamsID     uint // Standard field for the primary key
	Username    string
	Surname     string
	Email       string
	DateOfBirth datatypes.Date
	IsVegan     bool
	Occupation  occupation `gorm:"type:occupation"`
	CreatedAt   datatypes.Time
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
