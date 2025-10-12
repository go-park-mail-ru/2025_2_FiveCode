package authRepository

import (
	"backend/models"
	"backend/store"
	"github.com/pkg/errors"
)

type AuthRepository struct {
	Store *store.Store
}

func NewAuthRepository(store *store.Store) *AuthRepository {
	return &AuthRepository{Store: store}
}

func (r *AuthRepository) CreateSession(userID uint64) (string, error) {
	sessionID := r.Store.CreateSession(userID)
	return sessionID, nil
}

func (r *AuthRepository) GetUserByEmail(email string) (*models.User, error) {
	userID, ok := r.Store.UsersByEmail[email]
	if !ok {
		return nil, errors.New("user not found")
	}

	user, ok := r.Store.Users[userID]
	if !ok {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (r *AuthRepository) DeleteSession(sessionID string) error {
	r.Store.DeleteSession(sessionID)
	return nil
}
