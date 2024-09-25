package model

// Login response
type LoginResponse struct {
	TeamName string `json:"teamName"`
	Email    string `json:"email"`
}

// Struct represnting response for Get /team endpoint
type GetTeamResponse struct {
	TeamName    string                  `json:"teamName"`
	TeamMembers []GetTeamMemberResponse `json:"teamMembers"`
}

// Struct representing team member in response for Get /team endpoint
type GetTeamMemberResponse struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Verified  bool   `json:"verified"`
}
