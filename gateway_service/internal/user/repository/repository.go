package repository

import (
	"context"

	userPB "backend/user_service/pkg/user/v1"

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

func (r *UserRepository) GetUser(ctx context.Context, userID uint64) (*userPB.User, error) {
	return r.client.GetUser(ctx, &userPB.GetUserRequest{UserId: userID})
}

func (r *UserRepository) UpdateUser(ctx context.Context, req *userPB.UpdateUserRequest) (*userPB.User, error) {
	return r.client.UpdateUser(ctx, req)
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
