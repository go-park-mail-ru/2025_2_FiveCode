package server

import (
	"context"
	"errors"

	"backend/auth_service/internal/constants"
	pb "backend/auth_service/pkg/auth/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

//go:generate mockgen -source=delivery.go -destination=../mock/mock_delivery.go -package=mock
type AuthUsecase interface {
	CreateSession(ctx context.Context, userID uint64) (string, error)
	Logout(ctx context.Context, sessionID string) error
	GetUserIDBySession(ctx context.Context, sessionID string) (uint64, error)
	GenerateCSRFToken(ctx context.Context, sessionID string) (string, error)
}

type Server struct {
	pb.UnimplementedAuthServer

	Usecase AuthUsecase
}

func NewServer(uc AuthUsecase) *Server {
	return &Server{
		Usecase: uc,
	}
}

func RegisterService(grpcServer *grpc.Server, usecase AuthUsecase) {
	server := NewServer(usecase)
	pb.RegisterAuthServer(grpcServer, server)
}

func (s *Server) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	if req.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	sessionID, err := s.Usecase.CreateSession(ctx, req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create session")
	}

	return &pb.CreateSessionResponse{
		SessionId: sessionID,
	}, nil
}

func (s *Server) Logout(ctx context.Context, req *pb.LogoutRequest) (*emptypb.Empty, error) {
	err := s.Usecase.Logout(ctx, req.GetSessionId())
	if err != nil {
		if errors.Is(err, constants.ErrInvalidSession) {
			return &emptypb.Empty{}, nil
		}
		return nil, status.Error(codes.Internal, "failed to logout user")
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) GetUserIDBySession(ctx context.Context, req *pb.GetUserIDBySessionRequest) (*pb.GetUserIDBySessionResponse, error) {
	userID, err := s.Usecase.GetUserIDBySession(ctx, req.GetSessionId())

	if err != nil {
		if errors.Is(err, constants.ErrInvalidSession) {
			return &pb.GetUserIDBySessionResponse{
				UserId:  0,
				IsValid: false,
			}, nil
		}

		return nil, status.Error(codes.Internal, "failed to validate session")
	}

	return &pb.GetUserIDBySessionResponse{
		UserId:  userID,
		IsValid: true,
	}, nil
}

func (s *Server) GetCSRFToken(ctx context.Context, req *pb.GetCSRFTokenRequest) (*pb.GetCSRFTokenResponse, error) {
	token, err := s.Usecase.GenerateCSRFToken(ctx, req.GetSessionId())
	if err != nil {
		return nil, status.Error(codes.Internal, "could not generate csrf token")
	}

	return &pb.GetCSRFTokenResponse{Token: token}, nil
}
