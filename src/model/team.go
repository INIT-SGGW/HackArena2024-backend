package model

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Team struct {
	gorm.Model
	TeamName  string         // A regular string field
	Password  string         // A pointer to a string, allowing for null values
	CreatedAt datatypes.Time // Automatically managed by GORM for creation time
	UpdatedAt datatypes.Time // Automatically managed by GORM for update time
}
