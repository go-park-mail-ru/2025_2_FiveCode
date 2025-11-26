package usecase

import (
	"backend/user_service/internal/constants"
	"backend/user_service/internal/models"
	"backend/user_service/mock"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestUserUsecase_CreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockUserRepository(ctrl)
	usecase := NewUserUsecase(mockRepo)

	ctx := context.Background()
	email := "test@example.com"
	password := "password"
	username := "testuser"
	userID := uint64(1)
	now := time.Now()

	expectedUser := &models.User{
		ID:        userID,
		Email:     email,
		Username:  username,
		CreatedAt: now,
		UpdatedAt: &now,
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().
			CreateUser(ctx, email, gomock.Any(), username).
			Return(userID, nil)

		mockRepo.EXPECT().
			GetUserByID(ctx, userID).
			Return(expectedUser, nil)

		user, err := usecase.CreateUser(ctx, email, password, username)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
	})

	t.Run("CreateUser_Error", func(t *testing.T) {
		mockRepo.EXPECT().
			CreateUser(ctx, email, gomock.Any(), username).
			Return(uint64(0), errors.New("db error"))

		user, err := usecase.CreateUser(ctx, email, password, username)
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("GetUserByID_Error_After_Create", func(t *testing.T) {
		mockRepo.EXPECT().
			CreateUser(ctx, email, gomock.Any(), username).
			Return(userID, nil)

		mockRepo.EXPECT().
			GetUserByID(ctx, userID).
			Return(nil, errors.New("db error"))

		user, err := usecase.CreateUser(ctx, email, password, username)
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUserUsecase_GetUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockUserRepository(ctrl)
	usecase := NewUserUsecase(mockRepo)

	ctx := context.Background()
	userID := uint64(1)
	expectedUser := &models.User{
		ID:       userID,
		Username: "testuser",
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().
			GetUserByID(ctx, userID).
			Return(expectedUser, nil)

		user, err := usecase.GetUserByID(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo.EXPECT().
			GetUserByID(ctx, userID).
			Return(nil, errors.New("db error"))

		user, err := usecase.GetUserByID(ctx, userID)
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUserUsecase_UpdateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockUserRepository(ctrl)
	usecase := NewUserUsecase(mockRepo)

	ctx := context.Background()
	userID := uint64(1)
	newUsername := "newusername"
	newPassword := "newpassword"
	newAvatar := uint64(123)

	expectedUser := &models.User{
		ID:           userID,
		Username:     newUsername,
		AvatarFileID: &newAvatar,
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().
			UpdateUser(ctx, userID, &newUsername, gomock.Any(), &newAvatar).
			Return(expectedUser, nil)

		user, err := usecase.UpdateUser(ctx, userID, &newUsername, &newPassword, &newAvatar)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo.EXPECT().
			UpdateUser(ctx, userID, &newUsername, gomock.Any(), &newAvatar).
			Return(nil, errors.New("db error"))

		user, err := usecase.UpdateUser(ctx, userID, &newUsername, &newPassword, &newAvatar)
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUserUsecase_DeleteUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockUserRepository(ctrl)
	usecase := NewUserUsecase(mockRepo)

	ctx := context.Background()
	userID := uint64(1)

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().
			DeleteUser(ctx, userID).
			Return(nil)

		err := usecase.DeleteUser(ctx, userID)
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo.EXPECT().
			DeleteUser(ctx, userID).
			Return(errors.New("db error"))

		err := usecase.DeleteUser(ctx, userID)
		assert.Error(t, err)
	})
}

func TestUserUsecase_VerifyUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockUserRepository(ctrl)
	usecase := NewUserUsecase(mockRepo)

	ctx := context.Background()
	email := "test@example.com"
	password := "password"

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hashedPasswordStr := string(hashedPassword)

	expectedUser := &models.User{
		ID:       1,
		Email:    email,
		Password: hashedPasswordStr,
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().
			GetUserByEmail(ctx, email).
			Return(expectedUser, nil)

		user, err := usecase.VerifyUser(ctx, email, password)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.Empty(t, user.Password)
	})

	t.Run("WrongPassword", func(t *testing.T) {
		mockRepo.EXPECT().
			GetUserByEmail(ctx, email).
			Return(expectedUser, nil)

		user, err := usecase.VerifyUser(ctx, email, "wrongpassword")
		assert.Error(t, err)
		assert.Equal(t, constants.ErrInvalidEmailOrPassword, err)
		assert.Nil(t, user)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		mockRepo.EXPECT().
			GetUserByEmail(ctx, email).
			Return(nil, constants.ErrNotFound)

		user, err := usecase.VerifyUser(ctx, email, password)
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUserUsecase_GetUserByEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockUserRepository(ctrl)
	usecase := NewUserUsecase(mockRepo)

	ctx := context.Background()
	email := "test@example.com"
	expectedUser := &models.User{
		ID:    1,
		Email: email,
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo.EXPECT().
			GetUserByEmail(ctx, email).
			Return(expectedUser, nil)

		user, err := usecase.GetUserByEmail(ctx, email)
		assert.NoError(t, err)
		assert.Equal(t, expectedUser.ID, user.ID)
	})

	t.Run("Error", func(t *testing.T) {
		mockRepo.EXPECT().
			GetUserByEmail(ctx, email).
			Return(nil, errors.New("db error"))

		user, err := usecase.GetUserByEmail(ctx, email)
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}
