package usecase

import (
	"backend/pkg/logger"
	"backend/pkg/models"
	"backend/user/mock"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestUserUsecase_GetUserBySession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockAuthRepo := mock.NewMockAuthRepository(ctrl)
	usecase := NewUserUsecase(mockUserRepo, mockAuthRepo)

	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)

	tests := []struct {
		name          string
		sessionID     string
		setupMocks    func()
		expectedUser  *models.User
		expectedError error
	}{
		{
			name:      "success",
			sessionID: "test-session",
			setupMocks: func() {
				mockAuthRepo.EXPECT().GetUserIDBySession(ctx, "test-session").Return(uint64(1), nil)
				mockUserRepo.EXPECT().GetUserByID(ctx, uint64(1)).Return(&models.User{
					ID:        1,
					Email:     "test@example.com",
					Username:  "testuser",
					CreatedAt: time.Now(),
				}, nil)
			},
			expectedUser: &models.User{
				ID:        1,
				Email:     "test@example.com",
				Username:  "testuser",
				CreatedAt: time.Now(),
			},
			expectedError: nil,
		},
		{
			name:      "auth repo error",
			sessionID: "invalid-session",
			setupMocks: func() {
				mockAuthRepo.EXPECT().GetUserIDBySession(ctx, "invalid-session").Return(uint64(0), errors.New("session not found"))
			},
			expectedUser:  nil,
			expectedError: errors.New("failed to get user ID by session"),
		},
		{
			name:      "user repo error",
			sessionID: "test-session",
			setupMocks: func() {
				mockAuthRepo.EXPECT().GetUserIDBySession(ctx, "test-session").Return(uint64(1), nil)
				mockUserRepo.EXPECT().GetUserByID(ctx, uint64(1)).Return(nil, errors.New("user not found"))
			},
			expectedUser:  nil,
			expectedError: errors.New("failed to get user"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			user, err := usecase.GetUserBySession(ctx, tt.sessionID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser.ID, user.ID)
				assert.Equal(t, tt.expectedUser.Email, user.Email)
			}
		})
	}
}

func TestUserUsecase_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockAuthRepo := mock.NewMockAuthRepository(ctrl)
	usecase := NewUserUsecase(mockUserRepo, mockAuthRepo)

	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)

	tests := []struct {
		name          string
		username      *string
		password      *string
		avatarFileID  *uint64
		setupMocks    func()
		expectedUser  *models.User
		expectedError error
	}{
		{
			name:         "update username",
			username:     stringPtr("newusername"),
			password:     nil,
			avatarFileID: nil,
			setupMocks: func() {
				mockUserRepo.EXPECT().UpdateProfile(ctx, stringPtr("newusername"), gomock.Any(), nil).Return(&models.User{
					ID:        1,
					Username:  "newusername",
					CreatedAt: time.Now(),
				}, nil)
			},
			expectedUser: &models.User{
				ID:       1,
				Username: "newusername",
			},
			expectedError: nil,
		},
		{
			name:         "update password",
			username:     nil,
			password:     stringPtr("newpassword123"),
			avatarFileID: nil,
			setupMocks: func() {
				mockUserRepo.EXPECT().UpdateProfile(ctx, nil, gomock.Any(), nil).Return(&models.User{
					ID:        1,
					CreatedAt: time.Now(),
				}, nil)
			},
			expectedUser: &models.User{
				ID: 1,
			},
			expectedError: nil,
		},
		{
			name:         "update avatar",
			username:     nil,
			password:     nil,
			avatarFileID: uint64Ptr(10),
			setupMocks: func() {
				mockUserRepo.EXPECT().UpdateProfile(ctx, nil, nil, uint64Ptr(10)).Return(&models.User{
					ID:           1,
					AvatarFileID: uint64Ptr(10),
					CreatedAt:    time.Now(),
				}, nil)
			},
			expectedUser: &models.User{
				ID:           1,
				AvatarFileID: uint64Ptr(10),
			},
			expectedError: nil,
		},
		{
			name:         "repository error",
			username:     stringPtr("newusername"),
			password:     nil,
			avatarFileID: nil,
			setupMocks: func() {
				mockUserRepo.EXPECT().UpdateProfile(ctx, stringPtr("newusername"), gomock.Any(), nil).Return(nil, errors.New("database error"))
			},
			expectedUser:  nil,
			expectedError: errors.New("failed to update profile"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			user, err := usecase.UpdateProfile(ctx, tt.username, tt.password, tt.avatarFileID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.expectedUser.ID, user.ID)
			}
		})
	}
}

func TestUserUsecase_GetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockAuthRepo := mock.NewMockAuthRepository(ctrl)
	usecase := NewUserUsecase(mockUserRepo, mockAuthRepo)

	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)

	tests := []struct {
		name          string
		setupMocks    func()
		expectedUser  *models.User
		expectedError error
	}{
		{
			name: "success",
			setupMocks: func() {
				mockUserRepo.EXPECT().GetProfile(ctx).Return(&models.User{
					ID:        1,
					Email:     "test@example.com",
					Username:  "testuser",
					CreatedAt: time.Now(),
				}, nil)
			},
			expectedUser: &models.User{
				ID:       1,
				Email:    "test@example.com",
				Username: "testuser",
			},
			expectedError: nil,
		},
		{
			name: "repository error",
			setupMocks: func() {
				mockUserRepo.EXPECT().GetProfile(ctx).Return(nil, errors.New("database error"))
			},
			expectedUser:  nil,
			expectedError: errors.New("failed to get profile"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			user, err := usecase.GetProfile(ctx)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser.ID, user.ID)
				assert.Equal(t, tt.expectedUser.Email, user.Email)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}

func uint64Ptr(u uint64) *uint64 {
	return &u
}
