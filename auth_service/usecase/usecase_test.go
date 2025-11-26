package usecase

import (
	"backend/auth_service/mock"
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAuthUsecase_CreateSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockAuthRepository(ctrl)
	usecase := NewAuthUsecase(mockRepo, []byte("12345678901234567890123456789012"))

	ctx := context.Background()
	userID := uint64(1)
	sessionID := "session_123"

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().
			CreateSession(ctx, userID).
			Return(sessionID, nil)

		sid, err := usecase.CreateSession(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, sessionID, sid)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo.EXPECT().
			CreateSession(ctx, userID).
			Return("", errors.New("redis error"))

		sid, err := usecase.CreateSession(ctx, userID)
		assert.Error(t, err)
		assert.Empty(t, sid)
	})
}

func TestAuthUsecase_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockAuthRepository(ctrl)
	usecase := NewAuthUsecase(mockRepo, []byte("12345678901234567890123456789012"))

	ctx := context.Background()
	sessionID := "session_123"

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().
			DeleteSession(ctx, sessionID).
			Return(nil)

		err := usecase.Logout(ctx, sessionID)
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo.EXPECT().
			DeleteSession(ctx, sessionID).
			Return(errors.New("redis error"))

		err := usecase.Logout(ctx, sessionID)
		assert.Error(t, err)
	})
}

func TestAuthUsecase_GetUserIDBySession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockAuthRepository(ctrl)
	usecase := NewAuthUsecase(mockRepo, []byte("12345678901234567890123456789012"))

	ctx := context.Background()
	sessionID := "session_123"
	userID := uint64(1)

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().
			GetUserIDBySession(ctx, sessionID).
			Return(userID, nil)

		uid, err := usecase.GetUserIDBySession(ctx, sessionID)
		assert.NoError(t, err)
		assert.Equal(t, userID, uid)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo.EXPECT().
			GetUserIDBySession(ctx, sessionID).
			Return(uint64(0), errors.New("redis error"))

		uid, err := usecase.GetUserIDBySession(ctx, sessionID)
		assert.Error(t, err)
		assert.Zero(t, uid)
	})
}

func TestAuthUsecase_GenerateCSRFToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockAuthRepository(ctrl)
	usecase := NewAuthUsecase(mockRepo, []byte("12345678901234567890123456789012"))

	ctx := context.Background()
	sessionID := "session_123"

	t.Run("Success", func(t *testing.T) {
		token, err := usecase.GenerateCSRFToken(ctx, sessionID)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})
}
