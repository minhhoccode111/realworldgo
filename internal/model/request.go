package model

import (
	"fmt"
	"strings"

	"github.com/minhhoccode111/realworldgo/internal/utils"
)

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
	ac.Title, err = utils.IsValidTitle(ac.Title)
	if err != nil {
		return err
	}

	ac.Body = strings.TrimSpace(ac.Body)
	if ac.Body == "" {
		return fmt.Errorf("body cannot be empty")
	}

	ac.Description, err = utils.IsValidDescription(ac.Description)
	if err != nil {
		return err
	}

	ac.TagList, err = utils.IsValidTagList(ac.TagList)
	if err != nil {
		return err
	}
	return nil
}
