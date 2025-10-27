package userRepository

import (
	"backend/config"
	"backend/models"
	namederrors "backend/named_errors"
	"backend/store"
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type UserRepository struct {
	Store  *store.Store
	MinIO  *minio.Client
	Config *config.MinIOConfig
}

func NewUserRepository(store *store.Store, minioConfig *config.MinIOConfig) (*UserRepository, error) {
	minioClient, err := minio.New(minioConfig.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioConfig.AccessKeyID, minioConfig.SecretAccessKey, ""),
		Secure: minioConfig.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, minioConfig.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if !exists {
		err = minioClient.MakeBucket(ctx, minioConfig.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return &UserRepository{
		Store:  store,
		MinIO:  minioClient,
		Config: minioConfig,
	}, nil
}

func (r *UserRepository) CreateUser(email string, password string) (*models.User, error) {
	user, err := r.Store.CreateUser(email, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetUserBySession(sessionID string) (*models.User, error) {
	user, ok := r.Store.GetUserBySession(sessionID)
	if !ok {
		return nil, fmt.Errorf("failed to get user by session: %w", namederrors.ErrInvalidSession)
	}
	return user, nil
}

func (r *UserRepository) UpdateProfile(userID uint64, username *string, password *string) (*models.User, error) {
	user, err := r.Store.UpdateUserProfile(userID, username, password)
	if err != nil {
		return nil, namederrors.ErrUpdateProfile
	}
	return user, nil
}

func (r *UserRepository) GetProfile(userID uint64) (*models.User, error) {
	user, err := r.Store.GetUserByID(userID)
	if err != nil {
		return nil, namederrors.ErrGetProfile
	}
	return user, nil
}

func (r *UserRepository) SaveFile(file *models.File) (*models.File, error) {
	savedFile, err := r.Store.SaveFile(file)
	if err != nil {
		return nil, namederrors.ErrFileUpload
	}
	return savedFile, nil
}

func (r *UserRepository) UploadFileToMinIO(file io.Reader, filename, contentType string, size int64) (string, error) {
	objectName := fmt.Sprintf("%s-%s", uuid.New().String(), filename)

	ctx := context.Background()
	_, err := r.MinIO.PutObject(ctx, r.Config.BucketName, objectName, file, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to MinIO: %w", err)
	}

	url := fmt.Sprintf("http://%s/%s/%s", r.Config.Endpoint, r.Config.BucketName, objectName)
	return url, nil
}
