package usecase

import (
	"context"

	"backend/models"
	namederrors "backend/named_errors"
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
	UpdateProfile(ctx context.Context, userID uint64, username *string, password *string, avatarFileID *uint64) (*models.User, error)
	GetProfile(ctx context.Context, userID uint64) (*models.User, error)
	UploadAndSaveFile(ctx context.Context, file io.Reader, filename, contentType string, size int64, width, height int) (*models.File, error)
}

type AuthRepository interface {
	GetUserIDBySession(sessionID string) (uint64, error)
}

type UserUsecase struct {
	Repository   UserRepository
	AuthRepo     AuthRepository
}

func NewUserUsecase(UserRepository UserRepository, AuthRepo AuthRepository) *UserUsecase {
	return &UserUsecase{
		Repository: UserRepository,
		AuthRepo:   AuthRepo,
	}
}

func (uc *UserUsecase) RegisterUser(ctx context.Context, email string, password string) (*models.User, error) {
	user, err := uc.Repository.CreateUser(ctx, email, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (uc *UserUsecase) GetUserBySession(ctx context.Context, sessionID string) (*models.User, error) {
	userID, err := uc.AuthRepo.GetUserIDBySession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID by session: %w", err)
	}

	user, err := uc.Repository.GetProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	return user, nil
}

func (uc *UserUsecase) UpdateProfile(ctx context.Context, userID uint64, username *string, password *string, avatarFileID *uint64) (*models.User, error) {
	if password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		passwordStr := string(hashedPassword)
		password = &passwordStr
	}

	user, err := uc.Repository.UpdateProfile(ctx, userID, username, password, avatarFileID)
	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}
	return user, nil
}

func (uc *UserUsecase) GetProfile(ctx context.Context, userID uint64) (*models.User, error) {
	user, err := uc.Repository.GetProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}
	return user, nil
}

func (uc *UserUsecase) UploadAvatar(ctx context.Context, file io.Reader, filename, contentType string, size int64) (*models.File, error) {
	if contentType != "image/jpeg" && contentType != "image/png" {
		return nil, namederrors.ErrInvalidFileType
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	config, _, err := image.DecodeConfig(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image config: %w", err)
	}

	width := config.Width
	height := config.Height

	fileModel, err := uc.Repository.UploadAndSaveFile(ctx, bytes.NewReader(fileBytes), filename, contentType, size, width, height)
	if err != nil {
		return nil, err
	}

	return fileModel, nil
}
