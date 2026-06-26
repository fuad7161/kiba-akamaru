package model

import "time"

type User struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Email          string     `json:"email"`
	PasswordHash   string     `json:"-"`
	Role           string     `json:"role"`
	IsVerified     bool       `json:"is_verified"`
	VerifyToken    *string    `json:"-"`
	ResetToken     *string    `json:"-"`
	ResetTokenExp  *time.Time `json:"-"`
	Phone          *string    `json:"phone,omitempty"`
	District       *string    `json:"district,omitempty"`
	EducationLevel *string    `json:"education_level,omitempty"`
	LastLogin      *time.Time `json:"last_login,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// Public profile returned by GET /auth/me (no sensitive fields)
type UserProfile struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Role           string    `json:"role"`
	IsVerified     bool      `json:"is_verified"`
	Phone          *string   `json:"phone,omitempty"`
	District       *string   `json:"district,omitempty"`
	EducationLevel *string   `json:"education_level,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}
