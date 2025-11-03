package repository

import (
	"context"
	"fmt"
	"io"

	"backend/models"
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
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, userID uint64, username *string, password *string, avatarFileID *uint64) (*models.User, error) {
	user, err := r.Store.UpdateUserProfile(userID, username, password, avatarFileID)
	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetProfile(ctx context.Context, userID uint64) (*models.User, error) {
	user, err := r.Store.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}
	return user, nil
}

func (r *UserRepository) UploadAndSaveFile(ctx context.Context, file io.Reader, filename, contentType string, size int64, width, height int) (*models.File, error) {
	url, err := r.Store.UploadFileToMinIO(ctx, filename, file, size, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to MinIO: %w", err)
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
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	return savedFile, nil
}
