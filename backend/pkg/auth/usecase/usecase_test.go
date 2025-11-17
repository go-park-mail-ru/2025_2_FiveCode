package usecase

import (
	"backend/auth/mock"
	"backend/logger"
	"backend/models"
	namederrors "backend/named_errors"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthUsecase_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockAuthRepository(ctrl)
	usecase := NewAuthUsecase(mockRepo)

	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name          string
		email         string
		password      string
		setupMocks    func()
		expectedUser  *models.User
		expectedError error
	}{
		{
			name:    "success",
			email:   "test@example.com",
			password: "password123",
			setupMocks: func() {
				mockRepo.EXPECT().GetUserByEmail(ctx, "test@example.com").Return(&models.User{
					ID:        1,
					Email:     "test@example.com",
					Password:  string(hashedPassword),
					Username:  "test",
					CreatedAt: time.Now(),
				}, nil)
				mockRepo.EXPECT().CreateSession(ctx, uint64(1)).Return("session-id", nil)
			},
			expectedUser: &models.User{
				ID:       1,
				Email:    "test@example.com",
				Username: "test",
			},
			expectedError: nil,
		},
		{
			name:    "user not found",
			email:   "notfound@example.com",
			password: "password123",
			setupMocks: func() {
				mockRepo.EXPECT().GetUserByEmail(ctx, "notfound@example.com").Return(nil, namederrors.ErrNotFound)
			},
			expectedUser:  nil,
			expectedError: namederrors.ErrInvalidEmailOrPassword,
		},
		{
			name:    "invalid password",
			email:   "test@example.com",
			password: "wrongpassword",
			setupMocks: func() {
				mockRepo.EXPECT().GetUserByEmail(ctx, "test@example.com").Return(&models.User{
					ID:        1,
					Email:     "test@example.com",
					Password:  string(hashedPassword),
					Username:  "test",
					CreatedAt: time.Now(),
				}, nil)
			},
			expectedUser:  nil,
			expectedError: namederrors.ErrInvalidEmailOrPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			user, sessionID, err := usecase.Login(ctx, tt.email, tt.password)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, user)
				assert.Empty(t, sessionID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser.ID, user.ID)
				assert.Equal(t, tt.expectedUser.Email, user.Email)
				assert.NotEmpty(t, sessionID)
			}
		})
	}
}

func TestAuthUsecase_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockAuthRepository(ctrl)
	usecase := NewAuthUsecase(mockRepo)

	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)

	tests := []struct {
		name          string
		email         string
		password      string
		setupMocks    func()
		expectedUser  *models.User
		expectedError error
	}{
		{
			name:    "success",
			email:   "newuser@example.com",
			password: "password123",
			setupMocks: func() {
				mockRepo.EXPECT().CreateUser(ctx, "newuser@example.com", gomock.Any()).Return(&models.User{
					ID:        1,
					Email:     "newuser@example.com",
					Username:  "newuser",
					CreatedAt: time.Now(),
				}, nil)
				mockRepo.EXPECT().CreateSession(ctx, uint64(1)).Return("session-id", nil)
			},
			expectedUser: &models.User{
				ID:       1,
				Email:    "newuser@example.com",
				Username: "newuser",
			},
			expectedError: nil,
		},
		{
			name:    "user already exists",
			email:   "existing@example.com",
			password: "password123",
			setupMocks: func() {
				mockRepo.EXPECT().CreateUser(ctx, "existing@example.com", gomock.Any()).Return(nil, namederrors.ErrUserExists)
			},
			expectedUser:  nil,
			expectedError: namederrors.ErrUserExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			user, sessionID, err := usecase.Register(ctx, tt.email, tt.password)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, user)
				assert.Empty(t, sessionID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser.ID, user.ID)
				assert.Equal(t, tt.expectedUser.Email, user.Email)
				assert.NotEmpty(t, sessionID)
			}
		})
	}
}

func TestAuthUsecase_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockAuthRepository(ctrl)
	usecase := NewAuthUsecase(mockRepo)

	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)

	tests := []struct {
		name          string
		sessionID     string
		setupMocks    func()
		expectedError error
	}{
		{
			name:      "success",
			sessionID: "session-id",
			setupMocks: func() {
				mockRepo.EXPECT().DeleteSession(ctx, "session-id").Return(nil)
			},
			expectedError: nil,
		},
		{
			name:      "error deleting session",
			sessionID: "invalid-session",
			setupMocks: func() {
				mockRepo.EXPECT().DeleteSession(ctx, "invalid-session").Return(errors.New("session not found"))
			},
			expectedError: errors.New("failed to logout"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			err := usecase.Logout(ctx, tt.sessionID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthUsecase_GetUserBySession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockAuthRepository(ctrl)
	usecase := NewAuthUsecase(mockRepo)

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
			sessionID: "session-id",
			setupMocks: func() {
				mockRepo.EXPECT().GetUserIDBySession(ctx, "session-id").Return(uint64(1), nil)
				mockRepo.EXPECT().GetUserByID(ctx, uint64(1)).Return(&models.User{
					ID:        1,
					Email:     "test@example.com",
					Username:  "test",
					CreatedAt: time.Now(),
				}, nil)
			},
			expectedUser: &models.User{
				ID:       1,
				Email:    "test@example.com",
				Username: "test",
			},
			expectedError: nil,
		},
		{
			name:      "session not found",
			sessionID: "invalid-session",
			setupMocks: func() {
				mockRepo.EXPECT().GetUserIDBySession(ctx, "invalid-session").Return(uint64(0), namederrors.ErrInvalidSession)
			},
			expectedUser:  nil,
			expectedError: namederrors.ErrInvalidSession,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			user, err := usecase.GetUserBySession(ctx, tt.sessionID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser.ID, user.ID)
				assert.Equal(t, tt.expectedUser.Email, user.Email)
			}
		})
	}
}
