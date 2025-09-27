package model

import (
	"time"

	"github.com/minhhoccode111/realworldgo/internal/utils"
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

func (u *User) ValidateUserUpdateRequest(ur *UserUpdateRequest) error {
	var err error

	// ignore empty fields, if field is not empty, it must pass input validations
	// otherwise, reject whole process

	if ur.Username != "" {
		ur.Username, err = utils.IsValidUsername(ur.Username)
		u.Username = ur.Username
	}

	if ur.Email != "" {
		ur.Email, err = utils.IsValidEmail(ur.Email)
		u.Email = ur.Email
	}

	if ur.Password != "" {
		ur.Password, err = utils.IsValidPassword(ur.Password)
		u.Password = ur.Password
	}

	if ur.Bio != "" {
		u.Bio = ur.Bio
	}

	if ur.Image != "" {
		u.Image = ur.Image
	}

	return err
}
