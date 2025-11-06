package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const defaultBucketName = "notes-app"

type MinioStorage struct {
	client           *minio.Client
	internalEndpoint string // 0.0.0.0:8001
	publicEndpoint   string // http://89.208.210.115:8001
}

func NewMinioStorage(endpoint, accessKey, secretKey, publicEndpoint string, secure bool) (*MinioStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	storage := &MinioStorage{
		client:           client,
		internalEndpoint: endpoint,
		publicEndpoint:   publicEndpoint,
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

		policy := fmt.Sprintf(`{
            "Version": "2012-10-17",
            "Statement": [
                {
                    "Effect": "Allow",
                    "Principal": {"AWS": ["*"]},
                    "Action": ["s3:GetObject"],
                    "Resource": ["arn:aws:s3:::%s/*"]
                }
            ]
        }`, defaultBucketName)

		err = client.SetBucketPolicy(ctx, defaultBucketName, policy)
		if err != nil {
			return nil, fmt.Errorf("failed to set bucket policy: %w", err)
		}
	}

	return storage, nil
}

func (s *MinioStorage) GetClient() *minio.Client {
	return s.client
}

func (s *MinioStorage) TransformToPublicURL(internalURL string) string {
	if internalURL == "" {
		return ""
	}

	url := internalURL

	normalizedInternal := strings.Replace(s.internalEndpoint, "http://", "", 1)
	normalizedInternal = strings.Replace(normalizedInternal, "https://", "", 1)

	normalizedPublic := strings.Replace(s.publicEndpoint, "http://", "", 1)
	normalizedPublic = strings.Replace(normalizedPublic, "https://", "", 1)

	url = strings.Replace(url, normalizedInternal, normalizedPublic, 1)

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		if strings.HasPrefix(s.publicEndpoint, "https://") {
			url = "https://" + url
		} else {
			url = "http://" + url
		}
	}

	return url
}
