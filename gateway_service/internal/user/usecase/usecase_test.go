package usecase

import (
	"backend/gateway_service/internal/user/mock"
	"backend/gateway_service/internal/user/models"
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUserUsecase_GetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockAuthRepo := mock.NewMockAuthRepository(ctrl)
	usecase := NewUserUsecase(mockUserRepo, mockAuthRepo)

	ctx := context.Background()
	userID := uint64(1)
	expectedUser := &models.User{ID: userID, Email: "test@example.com"}

	t.Run("Success", func(t *testing.T) {
		mockUserRepo.EXPECT().
			GetUser(ctx, userID).
			Return(expectedUser, nil)

		user, err := usecase.GetProfile(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
	})

	t.Run("Error", func(t *testing.T) {
		mockUserRepo.EXPECT().
			GetUser(ctx, userID).
			Return(nil, errors.New("not found"))

		user, err := usecase.GetProfile(ctx, userID)
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUserUsecase_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockAuthRepo := mock.NewMockAuthRepository(ctrl)
	usecase := NewUserUsecase(mockUserRepo, mockAuthRepo)

	ctx := context.Background()
	input := &models.UpdateUserInput{ID: 1}
	expectedUser := &models.User{ID: 1, Username: "updated"}

	t.Run("Success", func(t *testing.T) {
		mockUserRepo.EXPECT().
			UpdateUser(ctx, input).
			Return(expectedUser, nil)

		user, err := usecase.UpdateProfile(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
	})
}

func TestUserUsecase_DeleteProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockAuthRepo := mock.NewMockAuthRepository(ctrl)
	usecase := NewUserUsecase(mockUserRepo, mockAuthRepo)

	ctx := context.Background()
	userID := uint64(1)
	sessionID := "session_123"

	t.Run("Success", func(t *testing.T) {
		mockUserRepo.EXPECT().
			DeleteUser(ctx, userID).
			Return(nil)

		mockAuthRepo.EXPECT().
			Logout(ctx, sessionID).
			Return(nil)

		err := usecase.DeleteProfile(ctx, userID, sessionID)
		assert.NoError(t, err)
	})

	t.Run("DeleteUser_Error", func(t *testing.T) {
		mockUserRepo.EXPECT().
			DeleteUser(ctx, userID).
			Return(errors.New("error"))

		err := usecase.DeleteProfile(ctx, userID, sessionID)
		assert.Error(t, err)
	})

	t.Run("Logout_Error", func(t *testing.T) {
		mockUserRepo.EXPECT().
			DeleteUser(ctx, userID).
			Return(nil)

		mockAuthRepo.EXPECT().
			Logout(ctx, sessionID).
			Return(errors.New("error"))

		err := usecase.DeleteProfile(ctx, userID, sessionID)
		assert.Error(t, err)
	})
}

func TestUserUsecase_GetProfileBySession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUserRepository(ctrl)
	mockAuthRepo := mock.NewMockAuthRepository(ctrl)
	usecase := NewUserUsecase(mockUserRepo, mockAuthRepo)

	ctx := context.Background()
	sessionID := "session_123"
	userID := uint64(1)
	expectedUser := &models.User{ID: userID}

	t.Run("Success", func(t *testing.T) {
		mockAuthRepo.EXPECT().
			GetUserIDBySession(ctx, sessionID).
			Return(userID, true, nil)

		mockUserRepo.EXPECT().
			GetUser(ctx, userID).
			Return(expectedUser, nil)

		user, err := usecase.GetProfileBySession(ctx, sessionID)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
	})

	t.Run("SessionInvalid", func(t *testing.T) {
		mockAuthRepo.EXPECT().
			GetUserIDBySession(ctx, sessionID).
			Return(uint64(0), false, nil)

		user, err := usecase.GetProfileBySession(ctx, sessionID)
		assert.NoError(t, err)
		assert.Nil(t, user)
	})

	t.Run("AuthError", func(t *testing.T) {
		mockAuthRepo.EXPECT().
			GetUserIDBySession(ctx, sessionID).
			Return(uint64(0), false, errors.New("error"))

		user, err := usecase.GetProfileBySession(ctx, sessionID)
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}
