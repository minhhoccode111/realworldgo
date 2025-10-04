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

type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserAuth struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Token    string `json:"token"`
	Bio      string `json:"bio"`
	Image    string `json:"image"`
}

func (u *User) ToUserAuth(token string) *UserAuth {
	return &UserAuth{
		Email:    u.Email,
		Username: u.Username,
		Token:    token,
		Bio:      u.Bio,
		Image:    u.Image,
	}
}

type ProfilePreview struct {
	Username  string `json:"username"`
	Bio       string `json:"bio"`
	Image     string `json:"image"`
	Following bool   `json:"following"`
}

func (u *User) ToProfilePreview(following bool) *ProfilePreview {
	return &ProfilePreview{
		Username:  u.Username,
		Bio:       u.Bio,
		Image:     u.Image,
		Following: following,
	}
}

type UserUpdate struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	Bio      string `json:"bio"`
	Image    string `json:"image"`
}

// ValidateUserUpdate validate and assign new values to user
func (u *User) ValidateUserUpdate(ur *UserUpdate) (err error) {
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

type UserRegister struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (ur *UserRegister) Validate() (err error) {
	ur.Username, err = utils.IsValidUsername(ur.Username)
	if err != nil {
		return err
	}

	ur.Email, err = utils.IsValidEmail(ur.Email)
	if err != nil {
		return err
	}

	ur.Password, err = utils.IsValidPassword(ur.Password)
	if err != nil {
		return err
	}
	return nil
}
