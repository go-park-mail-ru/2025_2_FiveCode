package usecase

import (
	"backend/gateway_service/internal/auth/mock"
	"backend/gateway_service/internal/user/models"
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAuthUsecase_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthRepo := mock.NewMockAuthRepository(ctrl)
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	usecase := NewAuthUsecase(mockAuthRepo, mockUserRepo)

	ctx := context.Background()
	email := "test@example.com"
	password := "password"
	userID := uint64(1)
	sessionID := "session_123"
	user := &models.User{ID: userID, Email: email}

	t.Run("Success", func(t *testing.T) {
		mockUserRepo.EXPECT().
			VerifyUser(ctx, email, password).
			Return(userID, nil)

		mockAuthRepo.EXPECT().
			CreateSession(ctx, userID).
			Return(sessionID, nil)

		mockUserRepo.EXPECT().
			GetUser(ctx, userID).
			Return(user, nil)

		sess, u, err := usecase.Login(ctx, email, password)
		assert.NoError(t, err)
		assert.Equal(t, sessionID, sess)
		assert.Equal(t, user, u)
	})

	t.Run("VerifyUser_Error", func(t *testing.T) {
		mockUserRepo.EXPECT().
			VerifyUser(ctx, email, password).
			Return(uint64(0), errors.New("invalid credentials"))

		sess, u, err := usecase.Login(ctx, email, password)
		assert.Error(t, err)
		assert.Empty(t, sess)
		assert.Nil(t, u)
	})

	t.Run("CreateSession_Error", func(t *testing.T) {
		mockUserRepo.EXPECT().
			VerifyUser(ctx, email, password).
			Return(userID, nil)

		mockAuthRepo.EXPECT().
			CreateSession(ctx, userID).
			Return("", errors.New("redis error"))

		sess, u, err := usecase.Login(ctx, email, password)
		assert.Error(t, err)
		assert.Empty(t, sess)
		assert.Nil(t, u)
	})

	t.Run("GetUser_Error", func(t *testing.T) {
		mockUserRepo.EXPECT().
			VerifyUser(ctx, email, password).
			Return(userID, nil)

		mockAuthRepo.EXPECT().
			CreateSession(ctx, userID).
			Return(sessionID, nil)

		mockUserRepo.EXPECT().
			GetUser(ctx, userID).
			Return(nil, errors.New("db error"))

		sess, u, err := usecase.Login(ctx, email, password)
		assert.Error(t, err)
		assert.Empty(t, sess)
		assert.Nil(t, u)
	})
}

func TestAuthUsecase_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthRepo := mock.NewMockAuthRepository(ctrl)
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	usecase := NewAuthUsecase(mockAuthRepo, mockUserRepo)

	ctx := context.Background()
	email := "test@example.com"
	password := "password"
	userID := uint64(1)
	sessionID := "session_123"
	user := &models.User{ID: userID, Email: email}

	t.Run("Success", func(t *testing.T) {
		mockUserRepo.EXPECT().
			CreateUser(ctx, email, password, "").
			Return(userID, nil)

		mockAuthRepo.EXPECT().
			CreateSession(ctx, userID).
			Return(sessionID, nil)

		mockUserRepo.EXPECT().
			GetUser(ctx, userID).
			Return(user, nil)

		sess, u, err := usecase.Register(ctx, email, password)
		assert.NoError(t, err)
		assert.Equal(t, sessionID, sess)
		assert.Equal(t, user, u)
	})

	t.Run("CreateUser_Error", func(t *testing.T) {
		mockUserRepo.EXPECT().
			CreateUser(ctx, email, password, "").
			Return(uint64(0), errors.New("user exists"))

		sess, u, err := usecase.Register(ctx, email, password)
		assert.Error(t, err)
		assert.Empty(t, sess)
		assert.Nil(t, u)
	})
}

func TestAuthUsecase_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthRepo := mock.NewMockAuthRepository(ctrl)
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	usecase := NewAuthUsecase(mockAuthRepo, mockUserRepo)

	ctx := context.Background()
	sessionID := "session_123"

	t.Run("Success", func(t *testing.T) {
		mockAuthRepo.EXPECT().
			Logout(ctx, sessionID).
			Return(nil)

		err := usecase.Logout(ctx, sessionID)
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		mockAuthRepo.EXPECT().
			Logout(ctx, sessionID).
			Return(errors.New("redis error"))

		err := usecase.Logout(ctx, sessionID)
		assert.Error(t, err)
	})
}

func TestAuthUsecase_GetCSRFToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthRepo := mock.NewMockAuthRepository(ctrl)
	mockUserRepo := mock.NewMockUserRepository(ctrl)
	usecase := NewAuthUsecase(mockAuthRepo, mockUserRepo)

	ctx := context.Background()
	sessionID := "session_123"
	token := "csrf_token"

	t.Run("Success", func(t *testing.T) {
		mockAuthRepo.EXPECT().
			GetCSRFToken(ctx, sessionID).
			Return(token, nil)

		tkn, err := usecase.GetCSRFToken(ctx, sessionID)
		assert.NoError(t, err)
		assert.Equal(t, token, tkn)
	})

	t.Run("Error", func(t *testing.T) {
		mockAuthRepo.EXPECT().
			GetCSRFToken(ctx, sessionID).
			Return("", errors.New("error"))

		tkn, err := usecase.GetCSRFToken(ctx, sessionID)
		assert.Error(t, err)
		assert.Empty(t, tkn)
	})
}
