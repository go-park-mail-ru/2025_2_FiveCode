package usecase

import (
	"backend/gateway_service/internal/file/mock"
	"backend/gateway_service/internal/file/models"
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestFileUsecase_UploadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockFileRepository(ctrl)
	usecase := NewFileUsecase(mockRepo)

	ctx := context.Background()
	filename := "test.txt"
	content := []byte("test content")
	fileReader := bytes.NewReader(content)
	contentType := "text/plain"
	size := int64(len(content))
	url := "http://minio/test.txt"

	fileModel := &models.File{
		ID:  1,
		URL: url,
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().
			UploadFileToMinIO(ctx, filename, content, contentType).
			Return(url, nil)

		mockRepo.EXPECT().
			SaveFile(ctx, url, contentType, size, gomock.Nil(), gomock.Nil()).
			Return(fileModel, nil)

		res, err := usecase.UploadFile(ctx, fileReader, filename, contentType, size)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, fileModel.ID, res.ID)
	})

	t.Run("UploadMinIO_Error", func(t *testing.T) {
		fileReader := bytes.NewReader(content)
		mockRepo.EXPECT().
			UploadFileToMinIO(ctx, filename, content, contentType).
			Return("", errors.New("minio error"))

		res, err := usecase.UploadFile(ctx, fileReader, filename, contentType, size)
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	t.Run("SaveFile_Error", func(t *testing.T) {
		fileReader := bytes.NewReader(content)
		mockRepo.EXPECT().
			UploadFileToMinIO(ctx, filename, content, contentType).
			Return(url, nil)

		mockRepo.EXPECT().
			SaveFile(ctx, url, contentType, size, gomock.Nil(), gomock.Nil()).
			Return(nil, errors.New("db error"))

		mockRepo.EXPECT().
			DeleteFileFromMinIO(ctx, url).
			Return(nil)

		res, err := usecase.UploadFile(ctx, fileReader, filename, contentType, size)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestFileUsecase_GetFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockFileRepository(ctrl)
	usecase := NewFileUsecase(mockRepo)

	ctx := context.Background()
	fileID := uint64(1)
	fileModel := &models.File{ID: fileID, URL: "http://minio/test.txt"}

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().
			GetFileByID(ctx, fileID).
			Return(fileModel, nil)

		res, err := usecase.GetFile(ctx, fileID)
		assert.NoError(t, err)
		assert.Equal(t, fileID, res.ID)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo.EXPECT().
			GetFileByID(ctx, fileID).
			Return(nil, errors.New("db error"))

		res, err := usecase.GetFile(ctx, fileID)
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestFileUsecase_DeleteFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockFileRepository(ctrl)
	usecase := NewFileUsecase(mockRepo)

	ctx := context.Background()
	fileID := uint64(1)
	fileModel := &models.File{ID: fileID, URL: "http://minio/test.txt"}

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().
			GetFileByID(ctx, fileID).
			Return(fileModel, nil)

		mockRepo.EXPECT().
			DeleteFileFromMinIO(ctx, fileModel.URL).
			Return(nil)

		mockRepo.EXPECT().
			DeleteFile(ctx, fileID).
			Return(nil)

		err := usecase.DeleteFile(ctx, fileID)
		assert.NoError(t, err)
	})

	t.Run("GetFile_Error", func(t *testing.T) {
		mockRepo.EXPECT().
			GetFileByID(ctx, fileID).
			Return(nil, errors.New("db error"))

		err := usecase.DeleteFile(ctx, fileID)
		assert.Error(t, err)
	})
}
