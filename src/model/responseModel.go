package model

import "gorm.io/datatypes"

// Login response
type LoginResponse struct {
	TeamName string `json:"teamName"`
	Email    string `json:"email"`
}

// Struct represnting response for Get /team endpoint
type GetTeamResponse struct {
	TeamName      string                  `json:"teamName"`
	IsVerified    bool                    `json:"verified"`
	ApproveStatus string                  `json:"approved"`
	IsConfirmed   bool                    `json:"confirmed"`
	TeamMembers   []GetTeamMemberResponse `json:"teamMembers"`
}

// Struct representing team member in response for Get /team endpoint
type GetTeamMemberResponse struct {
	Email     string  `json:"email"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Verified  bool    `json:"verified"`
}

// response wrapper for GET all teams admin endpoint
type GetAllTeamsResponse struct {
	Teams []TeamResponse `json:"teams"`
}

// Struct reporesenting team in response for GET all teams admin endpoints
type TeamResponse struct {
	TeamName         string `json:"teamName"`
	IsVerified       bool   `json:"verified"`
	ApproveSatatus   string `json:"approved"`
	IsConfirmed      bool   `json:"confirmed"`
	TeamMembersCount int    `json:"numberOfUsers"`
}

// response wrapper for Get all users admin endpoint
type GetAllUsersResponse struct {
	Users []UserResponse `json:"users"`
}

// struct representing team member in response for GET all users admin endpoints
type UserResponse struct {
	TeamName   string `json:"teamName"`
	Email      string `json:"email"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	IsVerified bool   `json:"verified"`
}

type UpdateTeamResponseBody struct {
	TeamName string `json:"teamName"`
}

type GetTeamMemberResponseBody struct {
	Email          string         `json:"email"`
	FirstName      string         `json:"firstName"`
	LastName       string         `json:"lastName"`
	DateOfBirth    datatypes.Date `json:"dateOfBirth"`
	Occupation     string         `json:"occupation"`
	DietPreference string         `json:"dietPreference"`
	School         string         `json:"school,omitempty"`
}
