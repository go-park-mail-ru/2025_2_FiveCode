package repository

import (
	"backend/auth_service/internal/constants"
	"context"
	"errors"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestAuthRepository_GetUserIDBySession(t *testing.T) {
	db, mockRedis := redismock.NewClientMock()
	repo := NewAuthRepository(db)
	ctx := context.Background()
	sessionID := "session-123"
	key := "session:" + sessionID
	userID := uint64(123)

	t.Run("Success", func(t *testing.T) {
		mockRedis.ExpectGet(key).SetVal("123")

		gotID, err := repo.GetUserIDBySession(ctx, sessionID)
		assert.NoError(t, err)
		assert.Equal(t, userID, gotID)
	})

	t.Run("NotFound", func(t *testing.T) {
		mockRedis.ExpectGet(key).RedisNil()

		gotID, err := repo.GetUserIDBySession(ctx, sessionID)
		assert.Error(t, err)
		assert.Equal(t, constants.ErrInvalidSession, err)
		assert.Equal(t, uint64(0), gotID)
	})

	t.Run("RedisError", func(t *testing.T) {
		mockRedis.ExpectGet(key).SetErr(errors.New("connection error"))

		gotID, err := repo.GetUserIDBySession(ctx, sessionID)
		assert.Error(t, err)
		assert.Equal(t, uint64(0), gotID)
	})

	t.Run("InvalidFormat", func(t *testing.T) {
		mockRedis.ExpectGet(key).SetVal("invalid-int")

		gotID, err := repo.GetUserIDBySession(ctx, sessionID)
		assert.Error(t, err)
		assert.Equal(t, uint64(0), gotID)
	})
}

func TestAuthRepository_DeleteSession(t *testing.T) {
	db, mockRedis := redismock.NewClientMock()
	repo := NewAuthRepository(db)
	ctx := context.Background()
	sessionID := "session-123"
	key := "session:" + sessionID

	t.Run("Success", func(t *testing.T) {
		mockRedis.ExpectDel(key).SetVal(1)

		err := repo.DeleteSession(ctx, sessionID)
		assert.NoError(t, err)
	})

	t.Run("RedisError", func(t *testing.T) {
		mockRedis.ExpectDel(key).SetErr(errors.New("connection error"))

		err := repo.DeleteSession(ctx, sessionID)
		assert.Error(t, err)
	})
}
