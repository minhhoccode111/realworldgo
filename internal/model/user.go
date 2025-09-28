package model

import (
	"time"

	"github.com/minhhoccode111/realworldgo/internal/utils"
)

type User struct {
	Id        string
	Email     string
	Username  string
	Image     string
	Bio       string
	CreatedAt time.Time
	UpdatedAt time.Time
	Password  string
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

func (u *User) ValidateUserUpdateRequest(ur *UserUpdateRequest) (err error) {
	// ignore empty fields, if field is not empty, it must pass input validations
	// otherwise, reject whole process

	if ur.Username != "" {
		u.Username, err = utils.IsValidUsername(ur.Username)
	}

	if ur.Email != "" {
		u.Email, err = utils.IsValidEmail(ur.Email)
	}

	if ur.Password != "" {
		u.Password, err = utils.IsValidPassword(ur.Password)
	}

	if ur.Bio != "" {
		u.Bio, err = utils.IsNotEmpty(ur.Bio)
	}

	if ur.Image != "" {
		u.Image, err = utils.IsNotEmpty(ur.Image)
	}

	return err
}
