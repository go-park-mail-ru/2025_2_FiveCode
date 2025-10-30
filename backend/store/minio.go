package store

import (
	"context"
	"errors"
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
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
	})
	if err != nil {
		return nil, errors.New("failed to create minio client:" + err.Error())
	}

	storage := &MinioStorage{
		client: client,
	}

	ctx := context.Background()

	exists, err := client.BucketExists(ctx, defaultBucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence (endpoint=%s, secure=%v): %w",
			endpoint, secure, err)
	}

	if !exists {
		err = client.MakeBucket(ctx, defaultBucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
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
