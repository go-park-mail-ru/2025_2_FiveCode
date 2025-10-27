package userUsecase

import (
	"backend/constants"
	"backend/models"
	namederrors "backend/named_errors"
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	CreateUser(email string, password string) (*models.User, error)
	GetUserBySession(sessionID string) (*models.User, error)
	UpdateProfile(userID uint64, username *string, password *string) (*models.User, error)
	GetProfile(userID uint64) (*models.User, error)
	SaveFile(file *models.File) (*models.File, error)
	UploadFileToMinIO(file io.Reader, filename, contentType string, size int64) (string, error)
}

type UserUsecase struct {
	Repository UserRepository
}

func NewUserUsecase(UserRepository UserRepository) *UserUsecase {
	return &UserUsecase{Repository: UserRepository}
}

func (uc *UserUsecase) RegisterUser(email string, password string) (*models.User, error) {
	user, err := uc.Repository.CreateUser(email, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (uc *UserUsecase) GetUserBySession(sessionID string) (*models.User, error) {
	user, err := uc.Repository.GetUserBySession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by session: %w", err)
	}
	return user, nil
}

func (uc *UserUsecase) UpdateProfile(userID uint64, username *string, password *string) (*models.User, error) {
	if username != nil {
		if len(*username) < 3 || len(*username) > 50 {
			return nil, namederrors.ErrUpdateProfile
		}
	}

	if password != nil {
		if len(*password) < 6 {
			return nil, namederrors.ErrUpdateProfile
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*password), bcrypt.DefaultCost)
		if err != nil {
			return nil, namederrors.ErrUpdateProfile
		}
		passwordStr := string(hashedPassword)
		password = &passwordStr
	}

	user, err := uc.Repository.UpdateProfile(userID, username, password)
	if err != nil {
		return nil, namederrors.ErrUpdateProfile
	}
	return user, nil
}

func (uc *UserUsecase) GetProfile(userID uint64) (*models.User, error) {
	user, err := uc.Repository.GetProfile(userID)
	if err != nil {
		return nil, namederrors.ErrGetProfile
	}
	return user, nil
}

func (uc *UserUsecase) UploadAvatar(file io.Reader, filename, contentType string, size int64) (*models.File, error) {
	if size > constants.MaxAvatarFileSize {
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

	url, err := uc.Repository.UploadFileToMinIO(bytes.NewReader(fileBytes), filename, contentType, size)
	if err != nil {
		return nil, err
	}

	fileModel := &models.File{
		ID:        0,
		URL:       url,
		MimeType:  contentType,
		SizeBytes: size,
		Width:     &width,
		Height:    &height,
		CreatedAt: time.Now().UTC(),
	}

	savedFile, err := uc.Repository.SaveFile(fileModel)
	if err != nil {
		return nil, err
	}

	return savedFile, nil
}
