package server

import (
	"backend/pkg/constants"
	"context"
	"errors"

	pb "backend/gen/go/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

//go:generate mockgen -source=delivery.go -destination=../mock/mock_delivery.go -package=mock
type AuthUsecase interface {
	Login(ctx context.Context, email string, password string) (uint64, string, error)
	Register(ctx context.Context, email string, password string) (uint64, string, error)
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

func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	userID, sessionId, err := s.Usecase.Register(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, constants.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user with this email already exists")
		}
		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &pb.RegisterResponse{
		UserId:    userID,
		SessionId: sessionId,
	}, nil
}

func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	userID, sessionId, err := s.Usecase.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, constants.ErrInvalidEmailOrPassword) {
			return nil, status.Error(codes.Unauthenticated, "invalid email or password")
		}
		return nil, status.Error(codes.Internal, "failed to login user")
	}

	return &pb.LoginResponse{
		UserId:    userID,
		SessionId: sessionId,
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
