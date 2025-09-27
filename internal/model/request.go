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

func (ur *UserRegisterRequest) Validate() (err error) {
	ur.Username, err = utils.IsValidUsername(ur.Username)
	ur.Email, err = utils.IsValidEmail(ur.Email)
	ur.Password, err = utils.IsValidPassword(ur.Password)
	return
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

func (ac *ArticleCreateRequest) Validate() (err error) {
	ac.Title, err = utils.IsNotEmpty(ac.Title)
	ac.Body, err = utils.IsNotEmpty(ac.Body)
	ac.Description, err = utils.IsNotEmpty(ac.Description)
	return
}
