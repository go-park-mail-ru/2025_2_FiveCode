package userRepository

import (
	"context"
	"errors"
	"io"

	"backend/models"
	namederrors "backend/named_errors"
	"backend/store"
)

type UserRepository struct {
	Store *store.Store
}

func NewUserRepository(store *store.Store) *UserRepository {
	return &UserRepository{
		Store: store,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, email string, password string) (*models.User, error) {
	user, err := r.Store.CreateUser(email, password)
	if err != nil {
		return nil, errors.New("failed to create user: " + err.Error())
	}
	return user, nil
}

func (r *UserRepository) GetUserBySession(ctx context.Context, sessionID string) (*models.User, error) {
	user, ok := r.Store.GetUserBySession(sessionID)
	if !ok {
		return nil, namederrors.ErrInvalidSession
	}
	return user, nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, userID uint64, username *string, password *string, avatarFileID *uint64) (*models.User, error) {
	user, err := r.Store.UpdateUserProfile(userID, username, password, avatarFileID)
	if err != nil {
		return nil, errors.New("failed to update profile: " + err.Error())
	}
	return user, nil
}

func (r *UserRepository) GetProfile(ctx context.Context, userID uint64) (*models.User, error) {
	user, err := r.Store.GetUserByID(userID)
	if err != nil {
		return nil, errors.New("failed to get profile: " + err.Error())
	}
	return user, nil
}

func (r *UserRepository) UploadAndSaveFile(ctx context.Context, file io.Reader, filename, contentType string, size int64, width, height int) (*models.File, error) {
	url, err := r.Store.UploadFileToMinIO(ctx, filename, file, size, contentType)
	if err != nil {
		return nil, errors.New("failed to upload file to MinIO: " + err.Error())
	}

	fileModel := &models.File{
		URL:       url,
		MimeType:  contentType,
		SizeBytes: size,
		Width:     &width,
		Height:    &height,
	}

	savedFile, err := r.Store.SaveFile(fileModel)
	if err != nil {
		return nil, errors.New("failed to save file: " + err.Error())
	}

	return savedFile, nil
}
