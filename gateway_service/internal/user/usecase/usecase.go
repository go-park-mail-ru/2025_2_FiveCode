package usecase

import (
	"backend/gateway_service/internal/user/models"
	"context"
)

type UserRepository interface {
	GetUser(ctx context.Context, userID uint64) (*models.User, error)
	UpdateUser(ctx context.Context, input *models.UpdateUserInput) (*models.User, error)
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

func (u *UserUsecase) GetProfile(ctx context.Context, userID uint64) (*models.User, error) {
	return u.userRepo.GetUser(ctx, userID)
}

func (u *UserUsecase) UpdateProfile(ctx context.Context, input *models.UpdateUserInput) (*models.User, error) {
	return u.userRepo.UpdateUser(ctx, input)
}

func (u *UserUsecase) DeleteProfile(ctx context.Context, userID uint64, sessionID string) error {
	if err := u.userRepo.DeleteUser(ctx, userID); err != nil {
		return err
	}

	if err := u.authRepo.Logout(ctx, sessionID); err != nil {
		return err
	}

	return nil
}

func (u *UserUsecase) GetProfileBySession(ctx context.Context, sessionID string) (*models.User, error) {
	userID, isValid, err := u.authRepo.GetUserIDBySession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if !isValid {
		return nil, nil
	}

	user, err := u.userRepo.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
