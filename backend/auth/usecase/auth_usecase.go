package authUsecase

import (
	"backend/models"
	"github.com/pkg/errors"
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
		return nil, "", errors.Wrap(err, "could not get user by email")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, "", errors.New("wrong password")
	}

	sessionID, err := uc.Repository.CreateSession(user.ID)
	if err != nil {
		return nil, "", errors.Wrap(err, "could not create session")
	}

	return user, sessionID, nil
}

func (uc *AuthUsecase) Logout(sessionID string) error {
	err := uc.Repository.DeleteSession(sessionID)
	if err != nil {
		return errors.Wrap(err, "could not delete session")
	}
	return nil
}
