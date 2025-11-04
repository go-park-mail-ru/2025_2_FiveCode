package Repository

import (
	"backend/logger"
	"backend/models"
	namederrors "backend/named_errors"
	"backend/store"
	"context"
)

type AuthRepository struct {
	Store *store.Store
}

func NewAuthRepository(store *store.Store) *AuthRepository {
	return &AuthRepository{Store: store}
}

func (r *AuthRepository) CreateSession(ctx context.Context, userID uint64) (string, error) {
	log := logger.FromContext(ctx)
	log.Info().Uint64("user_id", userID).Msg("creating session")
	sessionID := r.Store.CreateSession(userID)
	return sessionID, nil
}

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	log := logger.FromContext(ctx)
	log.Info().Str("email", email).Msg("getting user by email from store")
	userID, ok := r.Store.UsersByEmail[email]
	if !ok {
		log.Warn().Str("email", email).Msg("user not found by email")
		return nil, namederrors.ErrNotFound
	}

	user, ok := r.Store.Users[userID]
	if !ok {
		log.Error().Uint64("user_id", userID).Msg("user inconsistency: id found by email, but not in user map")
		return nil, namederrors.ErrNotFound
	}

	return user, nil
}

func (r *AuthRepository) DeleteSession(ctx context.Context, sessionID string) error {
	log := logger.FromContext(ctx)
	log.Info().Msg("deleting session from store")
	r.Store.DeleteSession(sessionID)
	return nil
}

func (r *AuthRepository) GetUserIDBySession(ctx context.Context, sessionID string) (uint64, error) {
	log := logger.FromContext(ctx)
	log.Info().Msg("getting user id by session from store")
	userID, ok := r.Store.GetUserIDBySession(ctx, sessionID)
	if !ok {
		log.Warn().Msg("session not found in store")
		return 0, namederrors.ErrInvalidSession
	}
	return userID, nil
}