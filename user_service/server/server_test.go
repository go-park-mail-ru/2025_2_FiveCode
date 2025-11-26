package server

import (
	"backend/user_service/internal/constants"
	"backend/user_service/internal/models"
	"backend/user_service/mock"
	user "backend/user_service/pkg/user/v1"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServer_CreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockUserUsecase(ctrl)
	server := NewServer(mockUsecase)
	ctx := context.Background()

	email := "test@example.com"
	password := "password"
	username := "testuser"
	now := time.Now()

	createdUser := &models.User{
		ID:        1,
		Email:     email,
		Username:  username,
		CreatedAt: now,
	}

	t.Run("Success", func(t *testing.T) {
		mockUsecase.EXPECT().
			CreateUser(ctx, email, password, username).
			Return(createdUser, nil)

		req := &user.CreateUserRequest{
			Email:    email,
			Password: password,
			Username: username,
		}
		resp, err := server.CreateUser(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, createdUser.ID, resp.Id)
		assert.Equal(t, createdUser.Email, resp.Email)
	})

	t.Run("UserExists", func(t *testing.T) {
		mockUsecase.EXPECT().
			CreateUser(ctx, email, password, username).
			Return(nil, constants.ErrUserExists)

		req := &user.CreateUserRequest{
			Email:    email,
			Password: password,
			Username: username,
		}
		resp, err := server.CreateUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.AlreadyExists, st.Code())
	})

	t.Run("InternalError", func(t *testing.T) {
		mockUsecase.EXPECT().
			CreateUser(ctx, email, password, username).
			Return(nil, errors.New("db error"))

		req := &user.CreateUserRequest{
			Email:    email,
			Password: password,
			Username: username,
		}
		resp, err := server.CreateUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Internal, st.Code())
	})
}

func TestServer_GetUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockUserUsecase(ctrl)
	server := NewServer(mockUsecase)
	ctx := context.Background()
	userID := uint64(1)
	now := time.Now()

	userModel := &models.User{
		ID:        userID,
		Email:     "test@example.com",
		Username:  "testuser",
		CreatedAt: now,
	}

	t.Run("Success", func(t *testing.T) {
		mockUsecase.EXPECT().
			GetUserByID(ctx, userID).
			Return(userModel, nil)

		resp, err := server.GetUser(ctx, &user.GetUserRequest{UserId: userID})
		assert.NoError(t, err)
		assert.Equal(t, userID, resp.Id)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockUsecase.EXPECT().
			GetUserByID(ctx, userID).
			Return(nil, constants.ErrNotFound)

		resp, err := server.GetUser(ctx, &user.GetUserRequest{UserId: userID})
		assert.Error(t, err)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.NotFound, st.Code())
	})

	t.Run("InternalError", func(t *testing.T) {
		mockUsecase.EXPECT().
			GetUserByID(ctx, userID).
			Return(nil, errors.New("db error"))

		resp, err := server.GetUser(ctx, &user.GetUserRequest{UserId: userID})
		assert.Error(t, err)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Internal, st.Code())
	})
}

func TestServer_UpdateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockUserUsecase(ctrl)
	server := NewServer(mockUsecase)
	ctx := context.Background()
	userID := uint64(1)
	username := "newname"
	avatarID := uint64(100)
	now := time.Now()

	userModel := &models.User{
		ID:           userID,
		Username:     username,
		AvatarFileID: &avatarID,
		CreatedAt:    now,
	}

	t.Run("Success", func(t *testing.T) {
		req := &user.UpdateUserRequest{
			UserId:       userID,
			Username:     username,
			AvatarFileId: avatarID,
		}

		mockUsecase.EXPECT().
			UpdateUser(ctx, userID, &username, nil, &avatarID).
			Return(userModel, nil)

		resp, err := server.UpdateUser(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, username, resp.Username)
		assert.Equal(t, avatarID, *resp.AvatarFileId)
	})

	t.Run("NotFound", func(t *testing.T) {
		req := &user.UpdateUserRequest{UserId: userID, Username: username}
		mockUsecase.EXPECT().
			UpdateUser(ctx, userID, &username, nil, nil).
			Return(nil, constants.ErrNotFound)

		resp, err := server.UpdateUser(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.NotFound, st.Code())
	})
}

func TestServer_DeleteUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockUserUsecase(ctrl)
	server := NewServer(mockUsecase)
	ctx := context.Background()
	userID := uint64(1)

	t.Run("Success", func(t *testing.T) {
		mockUsecase.EXPECT().
			DeleteUser(ctx, userID).
			Return(nil)

		resp, err := server.DeleteUser(ctx, &user.DeleteUserRequest{UserId: userID})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockUsecase.EXPECT().
			DeleteUser(ctx, userID).
			Return(constants.ErrNotFound)

		resp, err := server.DeleteUser(ctx, &user.DeleteUserRequest{UserId: userID})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("DeleteUser_NotFound_ReturnsSuccess", func(t *testing.T) {
		mockUsecase.EXPECT().
			DeleteUser(ctx, userID).
			Return(constants.ErrNotFound)

		resp, err := server.DeleteUser(ctx, &user.DeleteUserRequest{UserId: userID})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("InternalError", func(t *testing.T) {
		mockUsecase.EXPECT().
			DeleteUser(ctx, userID).
			Return(errors.New("db error"))

		resp, err := server.DeleteUser(ctx, &user.DeleteUserRequest{UserId: userID})
		assert.Error(t, err)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Internal, st.Code())
	})
}

func TestServer_VerifyUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockUserUsecase(ctrl)
	server := NewServer(mockUsecase)
	ctx := context.Background()
	email := "test@example.com"
	password := "password"
	userModel := &models.User{ID: 1, Email: email}

	t.Run("Success", func(t *testing.T) {
		mockUsecase.EXPECT().
			VerifyUser(ctx, email, password).
			Return(userModel, nil)

		resp, err := server.VerifyUser(ctx, &user.VerifyUserRequest{Email: email, Password: password})
		assert.NoError(t, err)
		assert.Equal(t, userModel.ID, resp.Id)
	})

	t.Run("Unauthenticated", func(t *testing.T) {
		mockUsecase.EXPECT().
			VerifyUser(ctx, email, password).
			Return(nil, errors.New("invalid"))

		resp, err := server.VerifyUser(ctx, &user.VerifyUserRequest{Email: email, Password: password})
		assert.Error(t, err)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})
}

func TestServer_GetUserByEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mock.NewMockUserUsecase(ctrl)
	server := NewServer(mockUsecase)
	ctx := context.Background()
	email := "test@example.com"
	userModel := &models.User{ID: 1, Email: email}

	t.Run("Success", func(t *testing.T) {
		mockUsecase.EXPECT().
			GetUserByEmail(ctx, email).
			Return(userModel, nil)

		resp, err := server.GetUserByEmail(ctx, &user.GetUserByEmailRequest{Email: email})
		assert.NoError(t, err)
		assert.Equal(t, userModel.ID, resp.Id)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockUsecase.EXPECT().
			GetUserByEmail(ctx, email).
			Return(nil, constants.ErrNotFound)

		resp, err := server.GetUserByEmail(ctx, &user.GetUserByEmailRequest{Email: email})
		assert.Error(t, err)
		assert.Nil(t, resp)
		st, _ := status.FromError(err)
		assert.Equal(t, codes.NotFound, st.Code())
	})
}

func TestRegisterService(t *testing.T) {
	s := grpc.NewServer()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUsecase := mock.NewMockUserUsecase(ctrl)
	RegisterService(s, mockUsecase)
}
