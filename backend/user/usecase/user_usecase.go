package userUsecase

import (
	"context"

	"backend/models"
	namederrors "backend/named_errors"
	"backend/validation"
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"

	"golang.org/x/crypto/bcrypt"
)

const (
	MaxAvatarFileSize = 10 * 1024 * 1024
)

type UserRepository interface {
	CreateUser(ctx context.Context, email string, password string) (*models.User, error)
	GetUserBySession(ctx context.Context, sessionID string) (*models.User, error)
	UpdateProfile(ctx context.Context, userID uint64, username *string, password *string, avatarFileID *uint64) (*models.User, error)
	GetProfile(ctx context.Context, userID uint64) (*models.User, error)
	UploadAndSaveFile(ctx context.Context, file io.Reader, filename, contentType string, size int64, width, height int) (*models.File, error)
}

type UserUsecase struct {
	Repository UserRepository
}

func NewUserUsecase(UserRepository UserRepository) *UserUsecase {
	return &UserUsecase{Repository: UserRepository}
}

func (uc *UserUsecase) RegisterUser(ctx context.Context, email string, password string) (*models.User, error) {
	user, err := uc.Repository.CreateUser(ctx, email, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (uc *UserUsecase) GetUserBySession(ctx context.Context, sessionID string) (*models.User, error) {
	user, err := uc.Repository.GetUserBySession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by session: %w", err)
	}
	return user, nil
}

type updatePasswordRequest struct {
	Password string `valid:"password"`
}

func (uc *UserUsecase) UpdateProfile(ctx context.Context, userID uint64, username *string, password *string, avatarFileID *uint64) (*models.User, error) {
	if username != nil {
		if len(*username) < 3 || len(*username) > 50 {
			return nil, namederrors.ErrUpdateProfile
		}
	}

	if password != nil {
		passwordReq := updatePasswordRequest{Password: *password}
		if err := validation.ValidateStruct(passwordReq); err != nil {
			return nil, namederrors.ErrUpdateProfile
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
		if err != nil {
			return nil, namederrors.ErrUpdateProfile
		}
		passwordStr := string(hashedPassword)
		password = &passwordStr
	}

	user, err := uc.Repository.UpdateProfile(ctx, userID, username, password, avatarFileID)
	if err != nil {
		return nil, namederrors.ErrUpdateProfile
	}
	return user, nil
}

func (uc *UserUsecase) GetProfile(ctx context.Context, userID uint64) (*models.User, error) {
	user, err := uc.Repository.GetProfile(ctx, userID)
	if err != nil {
		return nil, namederrors.ErrGetProfile
	}
	return user, nil
}

func (uc *UserUsecase) UploadAvatar(ctx context.Context, file io.Reader, filename, contentType string, size int64) (*models.File, error) {
	if size > MaxAvatarFileSize {
		return nil, namederrors.ErrFileTooLarge
	}

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
