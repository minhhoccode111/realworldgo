package model

import (
	"time"
)

type User struct {
	Id        string    `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Image     string    `json:"image"`
	Bio       string    `json:"bio"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Password  string    `json:"password"`
}

func (u *User) ToUserResponse(token string) *UserResponse {
	return &UserResponse{
		Email:    u.Email,
		Username: u.Username,
		Token:    token,
		Bio:      u.Bio,
		Image:    u.Image,
	}
}

type UserResponse struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Token    string `json:"token"`
	Bio      string `json:"bio"`
	Image    string `json:"image"`
}

type ProfileResponse struct {
	Username  string `json:"username"`
	Bio       string `json:"bio"`
	Image     string `json:"image"`
	Following bool   `json:"following"`
}
