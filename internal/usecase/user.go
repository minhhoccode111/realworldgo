package usecase

import "github.com/minhhoccode111/realworldgo/internal/entity"

type UserRepository interface {
	GetUserById(id string) (*entity.User, error)
}

type UserUsecase struct {
	userRepository UserRepository
}

func NewUserUsecase(r UserRepository) *UserUsecase {
	return &UserUsecase{
		userRepository: r,
	}
}

func (uu *UserUsecase) GetUser(id string) (*entity.User, error) {
	return uu.userRepository.GetUserById(id)
}
