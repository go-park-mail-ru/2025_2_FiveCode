package userUsecase

import (
	"backend/models"
	"github.com/pkg/errors"
)

type UserRepository interface {
	CreateUser(email string, password string) (*models.User, error)
	GetUserBySession(sessionID string) (*models.User, error)
}

type UserUsecase struct {
	Repository UserRepository
}

func NewUserUsecase(UserRepository UserRepository) *UserUsecase {
	return &UserUsecase{Repository: UserRepository}
}

func (uc *UserUsecase) RegisterUser(email string, password string) (*models.User, error) {
	user, err := uc.Repository.CreateUser(email, password)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create user")
	}

	return user, nil
}

func (uc *UserUsecase) GetUserBySession(sessionID string) (*models.User, error) {
	user, err := uc.Repository.GetUserBySession(sessionID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get user by session")
	}
	return user, nil
}
