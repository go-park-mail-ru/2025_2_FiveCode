package profileRepository

import (
	"backend/models"
	namederrors "backend/named_errors"
	"backend/store"
)

type ProfileRepository struct {
	Store *store.Store
}

func NewProfileRepository(store *store.Store) *ProfileRepository {
	return &ProfileRepository{
		Store: store,
	}
}

func (r *ProfileRepository) UpdateProfile(userID uint64, username *string) (*models.User, error) {
	user, err := r.Store.UpdateUserProfile(userID, username, nil)
	if err != nil {
		return nil, namederrors.ErrUpdateProfile
	}
	return user, nil
}

func (r *ProfileRepository) GetProfile(userID uint64) (*models.User, error) {
	user, err := r.Store.GetUserByID(userID)
	if err != nil {
		return nil, namederrors.ErrGetProfile
	}
	return user, nil
}

func (r *ProfileRepository) SaveFile(file *models.File) (*models.File, error) {
	savedFile, err := r.Store.SaveFile(file)
	if err != nil {
		return nil, namederrors.ErrFileUpload
	}
	return savedFile, nil
}
