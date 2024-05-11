package model

import "time"

type User struct {
	TeamsID   uint // Standard field for the primary key
	UserID    uint // Standard field for the primary key
	Username  string
	Surname   string
	Email     string
	CreatedAt time.Time
}
