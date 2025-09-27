package model

import "github.com/minhhoccode111/realworldgo/internal/utils"

type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserRegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (ur *UserRegisterRequest) Validate() error {
	var err error
	ur.Username, err = utils.IsValidUsername(ur.Username)
	ur.Email, err = utils.IsValidEmail(ur.Email)
	ur.Password, err = utils.IsValidPassword(ur.Password)
	return err
}

type UserUpdateRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	Bio      string `json:"bio"`
	Image    string `json:"image"`
}

type ArticleCreateRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Body        string   `json:"body"`
	TagList     []string `json:"tagList"`
}
