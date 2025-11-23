package repository

import (
	"backend/gateway_service/internal/user/models"
	"backend/gateway_service/internal/utils"
	userPB "backend/user_service/pkg/user/v1"
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserClient interface {
	GetUser(ctx context.Context, in *userPB.GetUserRequest, opts ...grpc.CallOption) (*userPB.User, error)
	UpdateUser(ctx context.Context, in *userPB.UpdateUserRequest, opts ...grpc.CallOption) (*userPB.User, error)
	DeleteUser(ctx context.Context, in *userPB.DeleteUserRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)

	VerifyUser(ctx context.Context, in *userPB.VerifyUserRequest, opts ...grpc.CallOption) (*userPB.User, error)
	CreateUser(ctx context.Context, in *userPB.CreateUserRequest, opts ...grpc.CallOption) (*userPB.User, error)
}

type UserRepository struct {
	client UserClient
}

func NewUserRepository(client UserClient) *UserRepository {
	return &UserRepository{
		client: client,
	}
}

func (r *UserRepository) GetUser(ctx context.Context, userID uint64) (*models.User, error) {
	resp, err := r.client.GetUser(ctx, &userPB.GetUserRequest{UserId: userID})
	if err != nil {
		return nil, err
	}
	return utils.MapProtoToUser(resp), nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, input *models.UpdateUserInput) (*models.User, error) {
	req := &userPB.UpdateUserRequest{
		UserId: input.ID,
	}
	if input.Username != nil {
		req.Username = *input.Username
	}
	if input.Password != nil {
		req.Password = *input.Password
	}
	if input.AvatarFileID != nil {
		req.AvatarFileId = *input.AvatarFileID
	}

	resp, err := r.client.UpdateUser(ctx, req)
	if err != nil {
		return nil, err
	}
	return utils.MapProtoToUser(resp), nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, userID uint64) error {
	_, err := r.client.DeleteUser(ctx, &userPB.DeleteUserRequest{UserId: userID})
	return err
}

func (r *UserRepository) VerifyUser(ctx context.Context, email, password string) (uint64, error) {
	resp, err := r.client.VerifyUser(ctx, &userPB.VerifyUserRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return 0, err
	}
	return resp.Id, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, email, password, username string) (uint64, error) {
	resp, err := r.client.CreateUser(ctx, &userPB.CreateUserRequest{
		Email:    email,
		Password: password,
		Username: username,
	})
	if err != nil {
		return 0, err
	}
	return resp.Id, nil
}
