package server

import (
	"backend/user_service/internal/constants"
	"backend/user_service/internal/models"
	"backend/user_service/logger"
	user "backend/user_service/pkg/user/v1"
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserUsecase interface {
	CreateUser(ctx context.Context, email, password, username string) (*models.User, error)
	GetUserByID(ctx context.Context, userID uint64) (*models.User, error)
	UpdateUser(ctx context.Context, userID uint64, username *string, password *string, avatarFileID *uint64) (*models.User, error)
	DeleteUser(ctx context.Context, userID uint64) error
	VerifyUser(ctx context.Context, email, password string) (*models.User, error)
}

type Server struct {
	user.UnimplementedUserServiceServer
	Usecase UserUsecase
}

func NewServer(uc UserUsecase) *Server {
	return &Server{
		Usecase: uc,
	}
}

func RegisterService(grpcServer *grpc.Server, usecase UserUsecase) {
	server := NewServer(usecase)
	user.RegisterUserServiceServer(grpcServer, server)
}

func (s *Server) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.User, error) {
	createdUser, err := s.Usecase.CreateUser(ctx, req.GetEmail(), req.GetPassword(), req.GetUsername())
	if err != nil {
		if errors.Is(err, constants.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user with this email already exists")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return modelUserToProto(createdUser), nil
}

func (s *Server) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.User, error) {
	log := logger.FromContext(ctx)
	log.Info().Uint64("user_id", req.GetUserId()).Msg("gRPC GetUser request")

	userModel, err := s.Usecase.GetUserByID(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		log.Error().Err(err).Msg("failed to get user")
		return nil, status.Error(codes.Internal, "failed to get user profile")
	}

	return modelUserToProto(userModel), nil
}

func (s *Server) UpdateUser(ctx context.Context, req *user.UpdateUserRequest) (*user.User, error) {
	log := logger.FromContext(ctx)
	log.Info().Uint64("user_id", req.GetUserId()).Msg("gRPC UpdateUser request")

	var username, password *string
	var avatarID *uint64

	if req.GetUsername() != "" {
		val := req.GetUsername()
		username = &val
	}
	if req.GetPassword() != "" {
		val := req.GetPassword()
		password = &val
	}
	if req.GetAvatarFileId() != 0 {
		val := req.GetAvatarFileId()
		avatarID = &val
	}

	updatedUser, err := s.Usecase.UpdateUser(ctx, req.GetUserId(), username, password, avatarID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		log.Error().Err(err).Msg("failed to update user")
		return nil, status.Error(codes.Internal, "failed to update user profile")
	}

	return modelUserToProto(updatedUser), nil
}

func (s *Server) DeleteUser(ctx context.Context, req *user.DeleteUserRequest) (*emptypb.Empty, error) {
	log := logger.FromContext(ctx)
	log.Info().Uint64("user_id", req.GetUserId()).Msg("gRPC DeleteUser request")

	err := s.Usecase.DeleteUser(ctx, req.GetUserId())
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return &emptypb.Empty{}, nil
		}
		log.Error().Err(err).Msg("failed to delete user")
		return nil, status.Error(codes.Internal, "failed to delete user profile")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) VerifyUser(ctx context.Context, req *user.VerifyUserRequest) (*user.User, error) {
	userModel, err := s.Usecase.VerifyUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid email or password")
	}

	return modelUserToProto(userModel), nil
}

func modelUserToProto(u *models.User) *user.User {
	if u == nil {
		return nil
	}
	protoUser := &user.User{
		Id:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		CreatedAt: timestamppb.New(u.CreatedAt),
	}
	if u.AvatarFileID != nil {
		protoUser.AvatarFileId = u.AvatarFileID
	}
	if u.UpdatedAt != nil {
		protoUser.UpdatedAt = timestamppb.New(*u.UpdatedAt)
	}
	return protoUser
}
