package model

import "time"

// Team object send to registration
type RegisterTeamRequest struct {
	TeamName          string   `json:"teamName" binding:"required"`
	TeamMembersEmails []string `json:"teamMembersEmails" binding:"required"`
}

// Team member struct for updating informaation about team member
type RegisterTeamMemberRequest struct {
	VerificationToken string    `json:"verificationToken" binding:"required"`
	Email             string    `json:"email" binding:"required"`
	Password          string    `json:"password" binding:"required"`
	FirstName         string    `json:"firstName" binding:"required"`
	LastName          string    `json:"lastName" binding:"required"`
	DateOfBirth       time.Time `json:"dateOfBirth" binding:"required"`
	Occupation        string    `json:"occupation" binding:"required"`
	DietPreference    string    `json:"dietPreference" binding:"required"`
	Agreement         bool      `json:"agreement" binding:"required"`
	School            string    `json:"school"` //preset only when occupation is "student"
}

// Login request input struct
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Standard change password request struct
type ChangePasswordRequest struct {
	NewPassword string `json:"newPassword" binding:"required"`
	OldPassword string `json:"oldPassword" binding:"required"`
}

// Reset password request
type ResetPasswordRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Token    string `json:"token" binding:"required"`
}

// Password restart request without the old password information
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required"`
}

// Admin register body
type RegisterAdminRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	UserName string `json:"userName" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginAdminRequest struct {
	User     string `json:"user" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdateTeamRequest struct {
	Status string `json:"status" binding:"required"`
}
