package usecase

import (
	"context"
	"fmt"

	userPB "backend/user_service/pkg/user/v1"
)

type UserRepository interface {
	GetUser(ctx context.Context, userID uint64) (*userPB.User, error)
	UpdateUser(ctx context.Context, req *userPB.UpdateUserRequest) (*userPB.User, error)
	DeleteUser(ctx context.Context, userID uint64) error
}

type AuthRepository interface {
	GetUserIDBySession(ctx context.Context, sessionID string) (uint64, bool, error)
	Logout(ctx context.Context, sessionID string) error
}

type UserUsecase struct {
	userRepo UserRepository
	authRepo AuthRepository
}

func NewUserUsecase(userRepo UserRepository, authRepo AuthRepository) *UserUsecase {
	return &UserUsecase{
		userRepo: userRepo,
		authRepo: authRepo,
	}
}

func (u *UserUsecase) GetProfile(ctx context.Context, userID uint64) (*userPB.User, error) {
	return u.userRepo.GetUser(ctx, userID)
}

func (u *UserUsecase) UpdateProfile(ctx context.Context, req *userPB.UpdateUserRequest) (*userPB.User, error) {
	return u.userRepo.UpdateUser(ctx, req)
}

func (u *UserUsecase) DeleteProfile(ctx context.Context, userID uint64, sessionID string) error {
	if err := u.userRepo.DeleteUser(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if err := u.authRepo.Logout(ctx, sessionID); err != nil {
		return fmt.Errorf("user deleted, but session logout failed: %w", err)
	}

	return nil
}

func (u *UserUsecase) GetProfileBySession(ctx context.Context, sessionID string) (*userPB.User, error) {
	userID, isValid, err := u.authRepo.GetUserIDBySession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to check session: %w", err)
	}

	if !isValid {
		return nil, nil
	}

	user, err := u.userRepo.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}
