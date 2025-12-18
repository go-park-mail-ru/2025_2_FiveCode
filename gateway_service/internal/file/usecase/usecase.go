package usecase

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"

	"backend/gateway_service/internal/constants"
	"backend/gateway_service/internal/file/models"
	"backend/gateway_service/internal/utils"
	"backend/pkg/logger"
)

//go:generate mockgen -source=usecase.go -destination=../mock/mock_usecase.go -package=mock
type FileRepository interface {
	UploadFileToMinIO(ctx context.Context, filename string, fileData []byte, contentType string) (string, error)
	SaveFile(ctx context.Context, url, mimeType string, sizeBytes int64, width, height *int) (*models.File, error)
	GetFileByID(ctx context.Context, fileID uint64) (*models.File, error)
	DeleteFile(ctx context.Context, fileID uint64) error
	DeleteFileFromMinIO(ctx context.Context, url string) error
	GetIcons(ctx context.Context) ([]*models.Icon, error)
	GetHeaders(ctx context.Context) ([]*models.Header, error)
}

type FileUsecase struct {
	Repository FileRepository
}

func NewFileUsecase(Repository FileRepository) *FileUsecase {
	return &FileUsecase{
		Repository: Repository,
	}
}

func (u *FileUsecase) UploadFile(ctx context.Context, file io.Reader, filename, contentType string, size int64) (*models.File, error) {
	log := logger.FromContext(ctx)

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Error().Err(err).Msg("failed to read file bytes")
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	detectedType := http.DetectContentType(fileBytes)
	if !isAllowedImageType(detectedType) {
		log.Warn().Str("detected_type", detectedType).Msg("upload rejected: invalid file type")
		return nil, fmt.Errorf("%w: %s, only images (jpeg, png, gif, webp, bmp) are allowed", constants.ErrInvalidFileType, detectedType)
	}

	_ = contentType
	contentType = detectedType

	var width, height *int
	config, _, err := image.DecodeConfig(bytes.NewReader(fileBytes))
	if err != nil {
		log.Warn().Err(err).Msg("failed to decode image config, proceeding without dimensions")
	} else {
		w := config.Width
		h := config.Height
		width = &w
		height = &h
		log.Info().Int("width", w).Int("height", h).Msg("decoded image dimensions")
	}

	url, err := u.Repository.UploadFileToMinIO(ctx, filename, fileBytes, contentType)
	if err != nil {
		log.Error().Err(err).Msg("failed to upload file to MinIO")
		return nil, fmt.Errorf("failed to upload file to MinIO: %w", err)
	}

	fileModel, err := u.Repository.SaveFile(ctx, url, contentType, size, width, height)
	if err != nil {
		log.Error().Err(err).Msg("failed to save file metadata")
		if deleteErr := u.Repository.DeleteFileFromMinIO(ctx, url); deleteErr != nil {
			log.Error().Err(deleteErr).Msg("failed to cleanup file from MinIO after metadata save failure")
		}
		return nil, fmt.Errorf("failed to save file metadata: %w", err)
	}

	fileModel.URL = utils.TransformMinioURL(fileModel.URL)

	return fileModel, nil
}

func (u *FileUsecase) GetFile(ctx context.Context, fileID uint64) (*models.File, error) {
	log := logger.FromContext(ctx)

	file, err := u.Repository.GetFileByID(ctx, fileID)
	if err != nil {
		log.Error().Err(err).Uint64("file_id", fileID).Msg("failed to get file")
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	file.URL = utils.TransformMinioURL(file.URL)

	return file, nil
}

func (u *FileUsecase) GetIcons(ctx context.Context) ([]models.Icon, error) {
	log := logger.FromContext(ctx)

	fileIcons, err := u.Repository.GetIcons(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get icons from repository")
		return nil, fmt.Errorf("failed to get icons: %w", err)
	}

	icons := make([]models.Icon, len(fileIcons))
	for i, fi := range fileIcons {
		icons[i] = models.Icon{
			ID:   fi.ID,
			Name: fi.Name,
			URL:  utils.TransformMinioURL(fi.URL),
		}
	}

	return icons, nil
}

func (u *FileUsecase) GetHeaders(ctx context.Context) ([]models.Header, error) {
	log := logger.FromContext(ctx)

	fileHeaders, err := u.Repository.GetHeaders(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get headers from repository")
		return nil, fmt.Errorf("failed to get headers: %w", err)
	}

	headers := make([]models.Header, len(fileHeaders))
	for i, fh := range fileHeaders {
		headers[i] = models.Header{
			ID:   fh.ID,
			Name: fh.Name,
			URL:  utils.TransformMinioURL(fh.URL),
		}
	}

	return headers, nil
}

func (u *FileUsecase) DeleteFile(ctx context.Context, fileID uint64) error {
	log := logger.FromContext(ctx)

	file, err := u.Repository.GetFileByID(ctx, fileID)
	if err != nil {
		log.Error().Err(err).Uint64("file_id", fileID).Msg("failed to get file for deletion")
		return fmt.Errorf("failed to get file: %w", err)
	}

	if err := u.Repository.DeleteFileFromMinIO(ctx, file.URL); err != nil {
		log.Error().Err(err).Str("url", file.URL).Msg("failed to delete file from MinIO")
	}

	if err := u.Repository.DeleteFile(ctx, fileID); err != nil {
		log.Error().Err(err).Uint64("file_id", fileID).Msg("failed to delete file from database")
		return fmt.Errorf("failed to delete file from database: %w", err)
	}

	return nil
}

func isAllowedImageType(contentType string) bool {
	return contentType == "image/jpeg" ||
		contentType == "image/png" ||
		contentType == "image/gif" ||
		contentType == "image/webp" ||
		contentType == "image/bmp"
}
