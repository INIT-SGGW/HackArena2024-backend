package model

import "time"

type Team struct {
	TeamsID   uint      // Standard field for the primary key
	TeamName  string    // A regular string field
	Password  string    // A pointer to a string, allowing for null values
	CreatedAt time.Time // Automatically managed by GORM for creation time
	UpdatedAt time.Time // Automatically managed by GORM for update time
}
