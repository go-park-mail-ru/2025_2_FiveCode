package usecase

import (
	"backend/logger"
	"backend/models"
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
)

//go:generate mockgen -source=usecase.go -destination=../mock/mock_usecase.go -package=mock
type FileRepository interface {
	UploadFileToMinIO(ctx context.Context, filename string, fileData []byte, contentType string) (string, error)
	SaveFile(ctx context.Context, url, mimeType string, sizeBytes int64, width, height *int) (*models.File, error)
	GetFileByID(ctx context.Context, fileID uint64) (*models.File, error)
	DeleteFile(ctx context.Context, fileID uint64) error
	DeleteFileFromMinIO(ctx context.Context, url string) error
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
	log.Info().Str("filename", filename).Str("content_type", contentType).Int64("size", size).Msg("uploading file")

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Error().Err(err).Msg("failed to read file bytes")
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var width, height *int
	if isImageContentType(contentType) {
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

	log.Info().Uint64("file_id", fileModel.ID).Msg("file uploaded successfully")
	return fileModel, nil
}

func (u *FileUsecase) GetFile(ctx context.Context, fileID uint64) (*models.File, error) {
	log := logger.FromContext(ctx)
	log.Info().Uint64("file_id", fileID).Msg("getting file")

	file, err := u.Repository.GetFileByID(ctx, fileID)
	if err != nil {
		log.Error().Err(err).Uint64("file_id", fileID).Msg("failed to get file")
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return file, nil
}

func (u *FileUsecase) DeleteFile(ctx context.Context, fileID uint64) error {
	log := logger.FromContext(ctx)
	log.Info().Uint64("file_id", fileID).Msg("deleting file")

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

	log.Info().Uint64("file_id", fileID).Msg("file deleted successfully")
	return nil
}

func isImageContentType(contentType string) bool {
	return contentType == "image/jpeg" ||
		contentType == "image/png" ||
		contentType == "image/gif" ||
		contentType == "image/webp"
}
