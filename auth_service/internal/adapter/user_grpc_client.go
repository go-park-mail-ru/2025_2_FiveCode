package adapter

import (
	"backend/auth_service/internal/constants"
	pb "backend/user_service/pkg/user/v1"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type UserGrpcAdapter struct {
	client pb.UserServiceClient
}

func NewUserGRPCAdapter(client pb.UserServiceClient) *UserGrpcAdapter {
	return &UserGrpcAdapter{client: client}
}

func ctxWithRequestID(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		vals := md.Get("request-id")
		if len(vals) > 0 {
			return metadata.AppendToOutgoingContext(ctx, "request-id", vals[0])
		}
	}
	return ctx
}

func (a *UserGrpcAdapter) VerifyUser(ctx context.Context, email, password string) (uint64, error) {
	outCtx := ctxWithRequestID(ctx)

	resp, err := a.client.VerifyUser(outCtx, &pb.VerifyUserRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		if status.Code(err) == codes.Unauthenticated {
			return 0, constants.ErrInvalidEmailOrPassword
		}
		return 0, err
	}
	return resp.GetId(), nil
}

func (a *UserGrpcAdapter) CreateUser(ctx context.Context, email, password string) (uint64, error) {
	outCtx := ctxWithRequestID(ctx)

	resp, err := a.client.CreateUser(outCtx, &pb.CreateUserRequest{
		Email:    email,
		Password: password,
		Username: "",
	})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return 0, constants.ErrUserExists
		}
		return 0, err
	}
	return resp.GetId(), nil
}
