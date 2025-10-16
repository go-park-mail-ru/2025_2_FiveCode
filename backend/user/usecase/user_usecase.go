package userUsecase

import (
	"backend/models"
	"fmt"
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
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (uc *UserUsecase) GetUserBySession(sessionID string) (*models.User, error) {
	user, err := uc.Repository.GetUserBySession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by session: %w", err)
	}
	return user, nil
}
