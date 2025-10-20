package profileUsecase

import (
	"backend/constants"
	"backend/models"
	namederrors "backend/named_errors"
	"io"
	"time"
)

type ProfileRepository interface {
	UpdateProfile(userID uint64, username *string) (*models.User, error)
	GetProfile(userID uint64) (*models.User, error)
	SaveFile(file *models.File) (*models.File, error)
}

type ProfileUsecase struct {
	Repository ProfileRepository
}

func NewProfileUsecase(repository ProfileRepository) *ProfileUsecase {
	return &ProfileUsecase{
		Repository: repository,
	}
}

func (uc *ProfileUsecase) UpdateProfile(userID uint64, username *string) (*models.User, error) {
	if username != nil {
		if len(*username) < 3 || len(*username) > 50 {
			return nil, namederrors.ErrUpdateProfile
		}
	}

	user, err := uc.Repository.UpdateProfile(userID, username)
	if err != nil {
		return nil, namederrors.ErrUpdateProfile
	}
	return user, nil
}

func (uc *ProfileUsecase) GetProfile(userID uint64) (*models.User, error) {
	user, err := uc.Repository.GetProfile(userID)
	if err != nil {
		return nil, namederrors.ErrGetProfile
	}
	return user, nil
}

func (uc *ProfileUsecase) UploadAvatar(file io.Reader, filename, contentType string, size int64) (*models.File, error) {
	if size > constants.MaxAvatarFileSize {
		return nil, namederrors.ErrFileTooLarge
	}

	if contentType != "image/jpeg" && contentType != "image/png" {
		return nil, namederrors.ErrInvalidFileType
	}

	fileModel := &models.File{
		ID:        0,
		URL:       "/api/files/file",
		MimeType:  contentType,
		SizeBytes: size,
		Width:     nil,
		Height:    nil,
		CreatedAt: time.Now().UTC(),
	}

	savedFile, err := uc.Repository.SaveFile(fileModel)
	if err != nil {
		return nil, err
	}

	return savedFile, nil
}
