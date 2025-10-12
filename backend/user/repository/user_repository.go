package userRepository

import (
	"backend/models"
	"backend/store"
	"github.com/pkg/errors"
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
		return nil, errors.Wrap(err, "create user failed")
	}

	return user, nil
}

func (r *UserRepository) GetUserBySession(sessionID string) (*models.User, error) {
	user, ok := r.Store.GetUserBySession(sessionID)
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}
