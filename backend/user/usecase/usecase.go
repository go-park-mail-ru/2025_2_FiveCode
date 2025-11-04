package usecase

import (
	"context"

	"backend/logger"
	"backend/models"
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	CreateUser(ctx context.Context, email string, password string) (*models.User, error)
	UpdateProfile(ctx context.Context, username *string, password *string, avatarFileID *uint64) (*models.User, error)
	GetProfile(ctx context.Context) (*models.User, error)
	GetUserByID(ctx context.Context, userID uint64) (*models.User, error)
	UploadAndSaveFile(ctx context.Context, file io.Reader, filename, contentType string, size int64, width, height int) (*models.File, error)
}

type AuthRepository interface {
	GetUserIDBySession(ctx context.Context, sessionID string) (uint64, error)
}

type UserUsecase struct {
	Repository UserRepository
	AuthRepo   AuthRepository
}

func NewUserUsecase(UserRepository UserRepository, AuthRepo AuthRepository) *UserUsecase {
	return &UserUsecase{
		Repository: UserRepository,
		AuthRepo:   AuthRepo,
	}
}

func (uc *UserUsecase) RegisterUser(ctx context.Context, email string, password string) (*models.User, error) {
	log := logger.FromContext(ctx)
	log.Info().Str("email", email).Msg("registering user")
	user, err := uc.Repository.CreateUser(ctx, email, password)
	if err != nil {
		log.Error().Err(err).Msg("failed to create user in repository")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (uc *UserUsecase) GetUserBySession(ctx context.Context, sessionID string) (*models.User, error) {
	log := logger.FromContext(ctx)
	log.Info().Msg("getting user by session")
	userID, err := uc.AuthRepo.GetUserIDBySession(ctx, sessionID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get user ID by session from auth repo")
		return nil, fmt.Errorf("failed to get user ID by session: %w", err)
	}

	user, err := uc.Repository.GetUserByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Uint64("user_id", userID).Msg("failed to get user profile from user repo")
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return user, nil
}

func (uc *UserUsecase) UpdateProfile(ctx context.Context, username *string, password *string, avatarFileID *uint64) (*models.User, error) {
	log := logger.FromContext(ctx)
	log.Info().Msg("updating user profile")
	if password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
		if err != nil {
			log.Error().Err(err).Msg("failed to hash password")
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		passwordStr := string(hashedPassword)
		password = &passwordStr
	}

	user, err := uc.Repository.UpdateProfile(ctx, username, password, avatarFileID)
	if err != nil {
		log.Error().Err(err).Msg("failed to update profile in repository")
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}
	return user, nil
}

func (uc *UserUsecase) GetProfile(ctx context.Context) (*models.User, error) {
	log := logger.FromContext(ctx)
	log.Info().Msg("getting user profile")
	user, err := uc.Repository.GetProfile(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get profile from repository")
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}
	return user, nil
}

func (uc *UserUsecase) UploadAvatar(ctx context.Context, file io.Reader, filename, contentType string, size int64) (*models.File, error) {
	log := logger.FromContext(ctx)
	log.Info().Str("filename", filename).Msg("uploading avatar")
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Error().Err(err).Msg("failed to read file bytes")
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	config, _, err := image.DecodeConfig(bytes.NewReader(fileBytes))
	if err != nil {
		log.Warn().Err(err).Msg("failed to decode image config, likely not an image")
		return nil, fmt.Errorf("failed to decode image config: %w", err)
	}

	width := config.Width
	height := config.Height
	log.Info().Int("width", width).Int("height", height).Msg("decoded image dimensions")

	fileModel, err := uc.Repository.UploadAndSaveFile(ctx, bytes.NewReader(fileBytes), filename, contentType, size, width, height)
	if err != nil {
		log.Error().Err(err).Msg("failed to upload and save file in repository")
		return nil, fmt.Errorf("failed to upload and save file: %w", err)
	}

	return fileModel, nil
}