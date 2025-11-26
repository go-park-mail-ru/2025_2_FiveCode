package repository

import (
	"backend/gateway_service/internal/user/models"
	"backend/gateway_service/internal/user/repository/mock"
	userPB "backend/user_service/pkg/user/v1"
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestUserRepository_GetUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockUserClient(ctrl)
	repo := NewUserRepository(mockClient)

	userID := uint64(1)
	protoUser := &userPB.User{Id: userID, Username: "test"}

	t.Run("Success", func(t *testing.T) {
		mockClient.EXPECT().GetUser(gomock.Any(), &userPB.GetUserRequest{UserId: userID}).Return(protoUser, nil)

		user, err := repo.GetUser(context.Background(), userID)
		assert.NoError(t, err)
		assert.Equal(t, userID, user.ID)
	})

	t.Run("Error", func(t *testing.T) {
		mockClient.EXPECT().GetUser(gomock.Any(), &userPB.GetUserRequest{UserId: userID}).Return(nil, errors.New("failed"))

		_, err := repo.GetUser(context.Background(), userID)
		assert.Error(t, err)
	})
}

func TestUserRepository_GetUserIDByEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockUserClient(ctrl)
	repo := NewUserRepository(mockClient)

	email := "test@example.com"
	protoUser := &userPB.User{Id: 1, Email: email}

	t.Run("Success", func(t *testing.T) {
		mockClient.EXPECT().GetUserByEmail(gomock.Any(), &userPB.GetUserByEmailRequest{Email: email}).Return(protoUser, nil)

		id, err := repo.GetUserIDByEmail(context.Background(), email)
		assert.NoError(t, err)
		assert.Equal(t, uint64(1), id)
	})
}

func TestUserRepository_UpdateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockUserClient(ctrl)
	repo := NewUserRepository(mockClient)

	userID := uint64(1)
	username := "updated"
	input := &models.UpdateUserInput{ID: userID, Username: &username}
	protoUser := &userPB.User{Id: userID, Username: username}

	t.Run("Success", func(t *testing.T) {
		mockClient.EXPECT().UpdateUser(gomock.Any(), &userPB.UpdateUserRequest{UserId: userID, Username: username}).Return(protoUser, nil)

		user, err := repo.UpdateUser(context.Background(), input)
		assert.NoError(t, err)
		assert.Equal(t, username, user.Username)
	})
}

func TestUserRepository_DeleteUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockUserClient(ctrl)
	repo := NewUserRepository(mockClient)

	userID := uint64(1)

	t.Run("Success", func(t *testing.T) {
		mockClient.EXPECT().DeleteUser(gomock.Any(), &userPB.DeleteUserRequest{UserId: userID}).Return(&emptypb.Empty{}, nil)

		err := repo.DeleteUser(context.Background(), userID)
		assert.NoError(t, err)
	})
}

func TestUserRepository_VerifyUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockUserClient(ctrl)
	repo := NewUserRepository(mockClient)

	email := "test@example.com"
	password := "pass"
	protoUser := &userPB.User{Id: 1}

	t.Run("Success", func(t *testing.T) {
		mockClient.EXPECT().VerifyUser(gomock.Any(), &userPB.VerifyUserRequest{Email: email, Password: password}).Return(protoUser, nil)

		id, err := repo.VerifyUser(context.Background(), email, password)
		assert.NoError(t, err)
		assert.Equal(t, uint64(1), id)
	})
}

func TestUserRepository_CreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mock.NewMockUserClient(ctrl)
	repo := NewUserRepository(mockClient)

	email := "test@example.com"
	password := "pass"
	username := "user"
	protoUser := &userPB.User{Id: 1}

	t.Run("Success", func(t *testing.T) {
		mockClient.EXPECT().CreateUser(gomock.Any(), &userPB.CreateUserRequest{Email: email, Password: password, Username: username}).Return(protoUser, nil)

		id, err := repo.CreateUser(context.Background(), email, password, username)
		assert.NoError(t, err)
		assert.Equal(t, uint64(1), id)
	})
}
