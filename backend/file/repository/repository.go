package repository

import (
	"backend/logger"
	"backend/models"
	namederrors "backend/named_errors"
	"backend/store"
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type FileRepository struct {
	Store *store.Store
}

func NewFileRepository(store *store.Store) *FileRepository {
	return &FileRepository{
		Store: store,
	}
}

func (r *FileRepository) UploadFileToMinIO(ctx context.Context, filename string, fileData []byte, contentType string) (string, error) {
	log := logger.FromContext(ctx)

	if r.Store.Minio == nil {
		return "", fmt.Errorf("minio storage not initialized")
	}

	objectName := fmt.Sprintf("%s-%s", uuid.New().String(), filename)
	bucketName := "notes-app"

	client := r.Store.Minio.GetClient()
	log.Info().Str("bucket", bucketName).Str("object", objectName).Msg("uploading file to MinIO")

	reader := bytes.NewReader(fileData)

	_, err := client.PutObject(ctx, bucketName, objectName,
		reader,
		int64(len(fileData)),
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to upload file to MinIO")
		return "", fmt.Errorf("failed to upload file to MinIO: %w", err)
	}

	endpoint := client.EndpointURL()
	url := fmt.Sprintf("%s://%s/%s/%s", endpoint.Scheme, endpoint.Host, bucketName, objectName)

	log.Info().Str("url", url).Msg("file uploaded to MinIO successfully")
	return url, nil
}

func (r *FileRepository) SaveFile(ctx context.Context, url, mimeType string, sizeBytes int64, width, height *int) (*models.File, error) {
	log := logger.FromContext(ctx)
	log.Info().Str("url", url).Msg("saving file metadata to PostgreSQL")

	query := `
		INSERT INTO file (url, mime_type, size_bytes, width, height)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, url, mime_type, size_bytes, width, height, created_at, updated_at
	`

	file := &models.File{}
	var widthResult, heightResult sql.NullInt32

	err := r.Store.Postgres.DB.QueryRowContext(ctx, query, url, mimeType, sizeBytes, width, height).Scan(
		&file.ID,
		&file.URL,
		&file.MimeType,
		&file.SizeBytes,
		&widthResult,
		&heightResult,
		&file.CreatedAt,
		&file.UpdatedAt,
	)

	if err != nil {
		log.Error().Err(err).Msg("failed to save file metadata")
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	if widthResult.Valid {
		w := int(widthResult.Int32)
		file.Width = &w
	}
	if heightResult.Valid {
		h := int(heightResult.Int32)
		file.Height = &h
	}

	log.Info().Uint64("file_id", file.ID).Msg("file metadata saved successfully")
	return file, nil
}

func (r *FileRepository) GetFileByID(ctx context.Context, fileID uint64) (*models.File, error) {
	log := logger.FromContext(ctx)
	log.Info().Uint64("file_id", fileID).Msg("getting file by id from PostgreSQL")

	query := `
		SELECT id, url, mime_type, size_bytes, width, height, created_at, updated_at
		FROM file
		WHERE id = $1
	`

	file := &models.File{}
	var width, height sql.NullInt32

	err := r.Store.Postgres.DB.QueryRowContext(ctx, query, fileID).Scan(
		&file.ID,
		&file.URL,
		&file.MimeType,
		&file.SizeBytes,
		&width,
		&height,
		&file.CreatedAt,
		&file.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		log.Warn().Uint64("file_id", fileID).Msg("file not found")
		return nil, namederrors.ErrNotFound
	}
	if err != nil {
		log.Error().Err(err).Uint64("file_id", fileID).Msg("failed to get file from PostgreSQL")
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	if width.Valid {
		w := int(width.Int32)
		file.Width = &w
	}
	if height.Valid {
		h := int(height.Int32)
		file.Height = &h
	}

	return file, nil
}

func (r *FileRepository) DeleteFile(ctx context.Context, fileID uint64) error {
	log := logger.FromContext(ctx)
	log.Info().Uint64("file_id", fileID).Msg("deleting file from PostgreSQL")

	query := `DELETE FROM file WHERE id = $1`

	result, err := r.Store.Postgres.DB.ExecContext(ctx, query, fileID)
	if err != nil {
		log.Error().Err(err).Uint64("file_id", fileID).Msg("failed to delete file")
		return fmt.Errorf("failed to delete file: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		log.Warn().Uint64("file_id", fileID).Msg("file not found for deletion")
		return namederrors.ErrNotFound
	}

	log.Info().Uint64("file_id", fileID).Msg("file deleted successfully")
	return nil
}

func (r *FileRepository) DeleteFileFromMinIO(ctx context.Context, url string) error {
	log := logger.FromContext(ctx)

	if r.Store.Minio == nil {
		return fmt.Errorf("minio storage not initialized")
	}

	objectName, err := extractObjectNameFromURL(url)
	if err != nil {
		log.Error().Err(err).Str("url", url).Msg("failed to extract object name from URL")
		return fmt.Errorf("invalid file URL: %w", err)
	}

	bucketName := "notes-app"
	client := r.Store.Minio.GetClient()

	log.Info().Str("bucket", bucketName).Str("object", objectName).Msg("deleting file from MinIO")

	err = client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		log.Error().Err(err).Msg("failed to delete file from MinIO")
		return fmt.Errorf("failed to delete file from MinIO: %w", err)
	}

	log.Info().Str("object", objectName).Msg("file deleted from MinIO successfully")
	return nil
}

func extractObjectNameFromURL(url string) (string, error) {
	parts := splitURL(url)
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid URL format")
	}
	return parts[len(parts)-1], nil
}

func splitURL(url string) []string {
	var parts []string
	current := ""
	for _, char := range url {
		if char == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
