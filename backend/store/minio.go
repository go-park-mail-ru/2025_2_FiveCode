package store

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
	"mime/multipart"
	"os"
)

const defaultBucketName = "notes-app"

type MinioStorage struct {
	client *minio.Client
}

func NewMinioStorage(endpoint, accessKey, secretKey string, secure bool) (*MinioStorage, error) {
	// Логируем параметры (БЕЗ secretKey!)
	log.Info().
		Str("endpoint", endpoint).
		Str("access_key", accessKey).
		Bool("secure", secure).
		Msg("Initializing MinIO client")

	// Создаем клиент
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	storage := &MinioStorage{
		client: client,
	}

	// Проверяем существование bucket
	ctx := context.Background()

	log.Info().Str("bucket", defaultBucketName).Msg("Checking bucket existence")

	exists, err := client.BucketExists(ctx, defaultBucketName)
	if err != nil {
		// Добавляем больше информации в ошибку
		return nil, fmt.Errorf("failed to check bucket existence (endpoint=%s, secure=%v): %w",
			endpoint, secure, err)
	}

	if !exists {
		log.Info().Str("bucket", defaultBucketName).Msg("Creating MinIO bucket")
		err = client.MakeBucket(ctx, defaultBucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Info().Str("bucket", defaultBucketName).Msg("MinIO bucket created successfully")
	} else {
		log.Info().Str("bucket", defaultBucketName).Msg("MinIO bucket already exists")
	}

	return storage, nil
}

func (s *MinioStorage) LoadImg(bucketName, fileName string, file multipart.File, fileSize int64) error {
	if bucketName == "" {
		bucketName = os.Getenv("MINIO_BUCKET_NAME")
	}
	_, err := s.client.PutObject(context.Background(), bucketName, fileName, file, fileSize, minio.PutObjectOptions{})
	return err
}

func (s *MinioStorage) DeleteImg(bucketName, fileName string) error {
	if bucketName == "" {
		bucketName = os.Getenv("MINIO_BUCKET_NAME")
	}
	return s.client.RemoveObject(context.Background(), bucketName, fileName, minio.RemoveObjectOptions{})
}

func (s *MinioStorage) GetFileURL(fileName string) string {
	endpoint := s.client.EndpointURL().String()
	return fmt.Sprintf("%s/%s/%s", endpoint, defaultBucketName, fileName)
}
