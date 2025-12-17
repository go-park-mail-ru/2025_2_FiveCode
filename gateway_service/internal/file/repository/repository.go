package repository

import (
	"backend/gateway_service/internal/constants"
	"backend/gateway_service/internal/file/models"
	"backend/pkg/logger"
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type FileRepository struct {
	db          *sql.DB
	minioClient *minio.Client
}

func NewFileRepository(db *sql.DB, minioClient *minio.Client) *FileRepository {
	return &FileRepository{
		db:          db,
		minioClient: minioClient,
	}
}

func (r *FileRepository) UploadFileToMinIO(ctx context.Context, filename string, fileData []byte, contentType string) (string, error) {
	log := logger.FromContext(ctx)

	objectName := fmt.Sprintf("%s-%s", uuid.New().String(), filename)
	bucketName := "notes-app"

	reader := bytes.NewReader(fileData)

	_, err := r.minioClient.PutObject(ctx, bucketName, objectName,
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

	endpoint := r.minioClient.EndpointURL()
	scheme := endpoint.Scheme
	if scheme == "" {
		scheme = "http"
	}

	internalURL := fmt.Sprintf("%s://%s/%s/%s", scheme, endpoint.Host, bucketName, objectName)

	return internalURL, nil
}

func (r *FileRepository) DeleteFileFromMinIO(ctx context.Context, url string) error {
	log := logger.FromContext(ctx)

	objectName, err := extractObjectNameFromURL(url)
	if err != nil {
		log.Error().Err(err).Str("url", url).Msg("failed to extract object name from URL")
		return fmt.Errorf("invalid file URL: %w", err)
	}

	bucketName := "notes-app"

	err = r.minioClient.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		log.Error().Err(err).Msg("failed to delete file from MinIO")
		return fmt.Errorf("failed to delete file from MinIO: %w", err)
	}

	return nil
}

func (r *FileRepository) SaveFile(ctx context.Context, url, mimeType string, sizeBytes int64, width, height *int) (*models.File, error) {
	log := logger.FromContext(ctx)

	now := time.Now().UTC()

	query := `
		INSERT INTO file (url, mime_type, size_bytes, width, height, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, url, mime_type, size_bytes, width, height, created_at, updated_at
	`

	file := &models.File{}
	var widthResult, heightResult sql.NullInt32

	err := r.db.QueryRowContext(ctx, query, url, mimeType, sizeBytes, width, height, now, now).Scan(
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

	return file, nil
}

func (r *FileRepository) GetFileByID(ctx context.Context, fileID uint64) (*models.File, error) {
	log := logger.FromContext(ctx)

	query := `
		SELECT id, url, mime_type, size_bytes, width, height, created_at, updated_at
		FROM file
		WHERE id = $1
	`

	file := &models.File{}
	var width, height sql.NullInt32

	err := r.db.QueryRowContext(ctx, query, fileID).Scan(
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
		return nil, constants.ErrNotFound
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

	query := `DELETE FROM file WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, fileID)
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
		return constants.ErrNotFound
	}

	return nil
}

func (r *FileRepository) GetIcons(ctx context.Context) ([]*models.Icon, error) {
	log := logger.FromContext(ctx)

	query := `
		SELECT id, url 
		FROM file 
		WHERE url LIKE '%/icons/%'
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.Error().Err(err).Msg("failed to query icons")
		return nil, fmt.Errorf("failed to query icons: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close rows")
		}
	}()

	var icons []*models.Icon
	for rows.Next() {
		var fileID uint64
		var url string
		if err := rows.Scan(&fileID, &url); err != nil {
			log.Error().Err(err).Msg("failed to scan icon row")
			return nil, fmt.Errorf("failed to scan icon: %w", err)
		}

		parts := strings.Split(url, "/")
		name := parts[len(parts)-1]

		icons = append(icons, &models.Icon{
			ID:   fileID,
			Name: name,
			URL:  url,
		})
	}

	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("error iterating icon rows")
		return nil, fmt.Errorf("failed to iterate icons: %w", err)
	}

	return icons, nil
}

func (r *FileRepository) GetHeaders(ctx context.Context) ([]*models.Header, error) {
	log := logger.FromContext(ctx)

	query := `
		SELECT id, url 
		FROM file 
		WHERE url LIKE '%/headers/%'
		ORDER BY id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		log.Error().Err(err).Msg("failed to query headers")
		return nil, fmt.Errorf("failed to query headers: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close rows")
		}
	}()

	var headers []*models.Header
	for rows.Next() {
		var fileID uint64
		var url string
		if err := rows.Scan(&fileID, &url); err != nil {
			log.Error().Err(err).Msg("failed to scan header row")
			return nil, fmt.Errorf("failed to scan header: %w", err)
		}

		parts := strings.Split(url, "/")
		name := parts[len(parts)-1]

		headers = append(headers, &models.Header{
			ID:   fileID,
			Name: name,
			URL:  url,
		})
	}

	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("error iterating header rows")
		return nil, fmt.Errorf("failed to iterate headers: %w", err)
	}

	return headers, nil
}

func extractObjectNameFromURL(url string) (string, error) {
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid URL format")
	}
	return parts[len(parts)-1], nil
}
