package usecase

import (
	"backend/file/mock"
	"backend/models"
	"context"
	"errors"
	"strings"
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

	tests := []struct {
		name          string
		contentType   string
		setupMocks    func()
		expectedError error
	}{
		{
			name:        "success",
			contentType: "text/plain",
			setupMocks: func() {
				mockRepo.EXPECT().UploadFileToMinIO(gomock.Any(), gomock.Any(), gomock.Any(), "text/plain").Return("http://example.com/file.txt", nil)
				mockRepo.EXPECT().SaveFile(gomock.Any(), "http://example.com/file.txt", "text/plain", int64(10), nil, nil).Return(&models.File{
					ID:        1,
					URL:       "http://example.com/file.txt",
					MimeType:  "text/plain",
					SizeBytes: 10,
				}, nil)
			},
			expectedError: nil,
		},
		{
			name:        "upload error",
			contentType: "text/plain",
			setupMocks: func() {
				mockRepo.EXPECT().UploadFileToMinIO(gomock.Any(), gomock.Any(), gomock.Any(), "text/plain").Return("", errors.New("upload failed"))
			},
			expectedError: errors.New("failed to upload file to MinIO"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			file := strings.NewReader("test content")
			result, err := usecase.UploadFile(ctx, file, "test.txt", tt.contentType, 10)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestFileUsecase_GetFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockFileRepository(ctrl)
	usecase := NewFileUsecase(mockRepo)

	ctx := context.Background()

	mockRepo.EXPECT().GetFileByID(ctx, uint64(1)).Return(&models.File{
		ID:        1,
		URL:       "http://example.com/file.txt",
		MimeType:  "text/plain",
		SizeBytes: 10,
	}, nil)

	file, err := usecase.GetFile(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), file.ID)
}

func TestFileUsecase_DeleteFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockFileRepository(ctrl)
	usecase := NewFileUsecase(mockRepo)

	ctx := context.Background()

	mockRepo.EXPECT().GetFileByID(ctx, uint64(1)).Return(&models.File{
		ID:  1,
		URL: "http://example.com/file.txt",
	}, nil)
	mockRepo.EXPECT().DeleteFileFromMinIO(ctx, "http://example.com/file.txt").Return(nil)
	mockRepo.EXPECT().DeleteFile(ctx, uint64(1)).Return(nil)

	err := usecase.DeleteFile(ctx, 1)
	assert.NoError(t, err)
}

func TestFileUsecase_UploadFile_WithImage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockFileRepository(ctrl)
	usecase := NewFileUsecase(mockRepo)

	ctx := context.Background()

	pngBytes := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xDE,
	}

	mockRepo.EXPECT().UploadFileToMinIO(gomock.Any(), gomock.Any(), gomock.Any(), "image/png").Return("http://example.com/image.png", nil)
	mockRepo.EXPECT().SaveFile(gomock.Any(), "http://example.com/image.png", "image/png", int64(len(pngBytes)), gomock.Any(), gomock.Any()).Return(&models.File{
		ID:        1,
		URL:       "http://example.com/image.png",
		MimeType:  "image/png",
		SizeBytes: int64(len(pngBytes)),
	}, nil)

	file := strings.NewReader(string(pngBytes))
	result, err := usecase.UploadFile(ctx, file, "test.png", "image/png", int64(len(pngBytes)))
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestFileUsecase_DeleteFile_ErrorGettingFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockFileRepository(ctrl)
	usecase := NewFileUsecase(mockRepo)

	ctx := context.Background()

	mockRepo.EXPECT().GetFileByID(ctx, uint64(1)).Return(nil, errors.New("file not found"))

	err := usecase.DeleteFile(ctx, 1)
	assert.Error(t, err)
}

func TestFileUsecase_DeleteFile_ErrorFromMinIO(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockFileRepository(ctrl)
	usecase := NewFileUsecase(mockRepo)

	ctx := context.Background()

	mockRepo.EXPECT().GetFileByID(ctx, uint64(1)).Return(&models.File{
		ID:  1,
		URL: "http://example.com/file.txt",
	}, nil)
	mockRepo.EXPECT().DeleteFileFromMinIO(ctx, "http://example.com/file.txt").Return(errors.New("minio error"))
	mockRepo.EXPECT().DeleteFile(ctx, uint64(1)).Return(nil)

	err := usecase.DeleteFile(ctx, 1)
	assert.NoError(t, err)
}

func TestFileUsecase_UploadFile_SaveFileError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockFileRepository(ctrl)
	usecase := NewFileUsecase(mockRepo)

	ctx := context.Background()

	mockRepo.EXPECT().UploadFileToMinIO(gomock.Any(), gomock.Any(), gomock.Any(), "text/plain").Return("http://example.com/file.txt", nil)
	mockRepo.EXPECT().SaveFile(gomock.Any(), "http://example.com/file.txt", "text/plain", int64(10), nil, nil).Return(nil, errors.New("save failed"))
	mockRepo.EXPECT().DeleteFileFromMinIO(ctx, "http://example.com/file.txt").Return(nil)

	file := strings.NewReader("test content")
	result, err := usecase.UploadFile(ctx, file, "test.txt", "text/plain", 10)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestFileUsecase_isImageContentType(t *testing.T) {
	assert.True(t, isImageContentType("image/jpeg"))
	assert.True(t, isImageContentType("image/png"))
	assert.True(t, isImageContentType("image/gif"))
	assert.True(t, isImageContentType("image/webp"))
	assert.False(t, isImageContentType("text/plain"))
	assert.False(t, isImageContentType("application/json"))
}
