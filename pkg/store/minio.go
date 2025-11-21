package store

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const defaultBucketName = "notes-app"

type MinioStorage struct {
	Client *minio.Client
}

func NewMinioStorage(endpoint, accessKey, secretKey string, secure bool) (*MinioStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	storage := &MinioStorage{
		Client: client,
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
