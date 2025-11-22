package repository

import (
	"backend/auth_service/internal/constants"
	"backend/auth_service/logger"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type AuthRepository struct {
	redis *redis.Client
}

func NewAuthRepository(redisClient *redis.Client) *AuthRepository {
	return &AuthRepository{
		redis: redisClient,
	}
}

func (r *AuthRepository) CreateSession(ctx context.Context, userID uint64) (string, error) {
	log := logger.FromContext(ctx)

	sessionID := uuid.NewString()
	key := "session:" + sessionID
	duration := 30 * 24 * time.Hour

	err := r.redis.Set(ctx, key, userID, duration).Err()
	if err != nil {
		log.Error().Err(err).Msg("failed to set session in redis")
		return "", err
	}

	return sessionID, nil
}

func (r *AuthRepository) GetUserIDBySession(ctx context.Context, sessionID string) (uint64, error) {
	log := logger.FromContext(ctx)

	key := "session:" + sessionID

	val, err := r.redis.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		log.Warn().Str("key", key).Msg("session not found in redis")
		return 0, constants.ErrInvalidSession
	}
	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("failed to get session from redis")
		return 0, err
	}

	var userID uint64
	_, err = fmt.Sscanf(val, "%d", &userID)
	if err != nil {
		log.Error().Err(err).Str("value", val).Msg("failed to parse userID from redis session value")
		return 0, fmt.Errorf("failed to parse userID from session: %w", err)
	}

	return userID, nil
}

func (r *AuthRepository) DeleteSession(ctx context.Context, sessionID string) error {
	log := logger.FromContext(ctx)

	key := "session:" + sessionID

	err := r.redis.Del(ctx, key).Err()
	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("failed to delete session from redis")
		return err
	}

	return nil
}
