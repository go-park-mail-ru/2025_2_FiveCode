package authUsecase

import (
	"backend/models"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type AuthRepository interface {
	GetUserByEmail(email string) (*models.User, error)
	CreateSession(userID uint64) (string, error)
	DeleteSession(sessionID string) error
}

type AuthUsecase struct {
	Repository AuthRepository
}

func NewAuthUsecase(repository AuthRepository) *AuthUsecase {
	return &AuthUsecase{Repository: repository}
}

func (uc *AuthUsecase) Login(email string, password string) (*models.User, string, error) {
	user, err := uc.Repository.GetUserByEmail(email)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get user by email: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, "", fmt.Errorf("wrong password: %w", err)
	}

	sessionID, err := uc.Repository.CreateSession(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create session: %w", err)
	}

	return user, sessionID, nil
}

func (uc *AuthUsecase) Logout(sessionID string) error {
	err := uc.Repository.DeleteSession(sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}
