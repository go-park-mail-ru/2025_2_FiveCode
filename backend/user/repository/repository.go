package repository

import (
	"context"
	"fmt"
	"io"

	"backend/logger"
	"backend/middleware"
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
	log := logger.FromContext(ctx)
	log.Info().Str("email", email).Msg("creating user in store")
	user, err := r.Store.CreateUser(email, password)
	if err != nil {
		log.Error().Err(err).Str("email", email).Msg("failed to create user in store")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) UpdateProfile(ctx context.Context, username *string, password *string, avatarFileID *uint64) (*models.User, error) {
	log := logger.FromContext(ctx)
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		log.Error().Msg("user not authenticated in repository layer")
		return nil, fmt.Errorf("user not authenticated")
	}
	log.Info().Uint64("user_id", userID).Msg("updating user profile in store")
	user, err := r.Store.UpdateUserProfile(userID, username, password, avatarFileID)
	if err != nil {
		log.Error().Err(err).Uint64("user_id", userID).Msg("failed to update profile in store")
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetProfile(ctx context.Context) (*models.User, error) {
	log := logger.FromContext(ctx)
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		log.Error().Msg("user not authenticated in repository layer")
		return nil, fmt.Errorf("user not authenticated")
	}
	log.Info().Uint64("user_id", userID).Msg("getting user profile from store")
	user, err := r.Store.GetUserByID(userID)
	if err != nil {
		log.Error().Err(err).Uint64("user_id", userID).Msg("failed to get profile from store")
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, userID uint64) (*models.User, error) {
	log := logger.FromContext(ctx)
	log.Info().Uint64("user_id", userID).Msg("getting user by id from store")
	user, err := r.Store.GetUserByID(userID)
	if err != nil {
		log.Error().Err(err).Uint64("user_id", userID).Msg("failed to get user from store")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (r *UserRepository) UploadAndSaveFile(ctx context.Context, file io.Reader, filename, contentType string, size int64, width, height int) (*models.File, error) {
	log := logger.FromContext(ctx)
	log.Info().Str("filename", filename).Msg("uploading file to minio")
	url, err := r.Store.UploadFileToMinIO(ctx, filename, file, size, contentType)
	if err != nil {
		log.Error().Err(err).Msg("failed to upload file to MinIO")
		return nil, fmt.Errorf("failed to upload file to MinIO: %w", err)
	}

	fileModel := &models.File{
		URL:       url,
		MimeType:  contentType,
		SizeBytes: size,
		Width:     &width,
		Height:    &height,
	}

	log.Info().Str("filename", filename).Msg("saving file metadata to store")
	savedFile, err := r.Store.SaveFile(fileModel)
	if err != nil {
		log.Error().Err(err).Msg("failed to save file metadata")
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	return savedFile, nil
}