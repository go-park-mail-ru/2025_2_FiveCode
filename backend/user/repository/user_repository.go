package userRepository

import (
	"backend/models"
	namederrors "backend/named_errors"
	"backend/store"
	"fmt"
)

type UserRepository struct {
	Store *store.Store
}

func NewUserRepository(store *store.Store) *UserRepository {
	return &UserRepository{
		Store: store,
	}
}

func (r *UserRepository) CreateUser(email string, password string) (*models.User, error) {
	user, err := r.Store.CreateUser(email, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetUserBySession(sessionID string) (*models.User, error) {
	user, ok := r.Store.GetUserBySession(sessionID)
	if !ok {
		return nil, fmt.Errorf("failed to get user by session: %w", namederrors.ErrInvalidSession)
	}
	return user, nil
}
