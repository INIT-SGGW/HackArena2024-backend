package model

// Login response
type LoginResponse struct {
	TeamName string `json:"teamName"`
	Email    string `json:"email"`
}

// Struct represnting response for Get /team endpoint
type GetTeamResponse struct {
	TeamName    string                  `json:"teamName"`
	IsVerified  bool                    `json:"verified"`
	TeamMembers []GetTeamMemberResponse `json:"teamMembers"`
}

// Struct representing team member in response for Get /team endpoint
type GetTeamMemberResponse struct {
	Email     string  `json:"email"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Verified  bool    `json:"verified"`
}

// response wrapper for Get all teams admin endpoint
type GetAllTeamsResponse struct {
	Teams []TeamResponse `json:"teams"`
}

// Struct reporesenting team in response fro Get all teams admin endpoints
type TeamResponse struct {
	TeamName         string `json:"teamName"`
	IsVerified       bool   `json:"verified"`
	ApproveSatatus   string `json:"approved"`
	TeamMembersCount int    `json:"numberOfUsers"`
}
