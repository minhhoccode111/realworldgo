package model

import (
	"strings"
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

func (u *User) ToProfilePreviewResponse(following bool) *ProfilePreviewResponse {
	return &ProfilePreviewResponse{
		Username:  u.Username,
		Bio:       u.Bio,
		Image:     u.Image,
		Following: following,
	}
}

func (u *User) ValidateUserUpdateRequest(ur *UserUpdateRequest) (err error) {
	// ignore empty fields, if field is not empty, it must pass input validations
	// otherwise, reject whole process

	if ur.Username != "" {
		u.Username, err = utils.IsValidUsername(ur.Username)
		if err != nil {
			return err
		}
	}

	if ur.Email != "" {
		u.Email, err = utils.IsValidEmail(ur.Email)
		if err != nil {
			return err
		}
	}

	if ur.Password != "" {
		u.Password, err = utils.IsValidPassword(ur.Password)
		if err != nil {
			return err
		}
	}

	if ur.Bio != "" {
		u.Bio = strings.TrimSpace(u.Bio)
	}

	if ur.Image != "" {
		u.Image = strings.TrimSpace(u.Image)
	}

	return nil
}
