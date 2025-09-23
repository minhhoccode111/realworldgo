package handler

import "github.com/minhhoccode111/realworldgo/internal/usecase"

type UserHandler struct {
	uc *usecase.UserUsecase
}

func NewUserHandler(uc *usecase.UserUsecase) *UserHandler {
	return &UserHandler{
		uc: uc,
	}
}
