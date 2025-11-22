package repository

import (
	"backend/gateway_service/internal/middleware"
	"backend/pkg/store"
	"backend/user_service/logger"
	"backend/user_service/models"
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *store.Store) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}

	store := &store.Store{
		Postgres: &store.PostgresDB{DB: db},
	}

	return db, mock, store
}

func TestUserRepository_GetProfile(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(store)
	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)
	ctx = middleware.WithUserID(ctx, 1)

	tests := []struct {
		name          string
		setupMocks    func()
		expectedUser  *models.User
		expectedError error
	}{
		{
			name: "success",
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "username", "avatar_file_id", "created_at", "updated_at"}).
					AddRow(1, "test@example.com", "hash", "testuser", nil, time.Now(), nil)
				mock.ExpectQuery(`SELECT id, email, password_hash, username, avatar_file_id, created_at, updated_at`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedUser: &models.User{
				ID:       1,
				Email:    "test@example.com",
				Username: "testuser",
			},
			expectedError: nil,
		},
		{
			name: "user not found",
			setupMocks: func() {
				mock.ExpectQuery(`SELECT id, email, password_hash, username, avatar_file_id, created_at, updated_at`).
					WithArgs(1).
					WillReturnError(sql.ErrNoRows)
			},
			expectedUser:  nil,
			expectedError: namederrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			user, err := repo.GetProfile(ctx)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser.ID, user.ID)
				assert.Equal(t, tt.expectedUser.Email, user.Email)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_GetUserByID(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(store)
	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)

	tests := []struct {
		name          string
		userID        uint64
		setupMocks    func()
		expectedUser  *models.User
		expectedError error
	}{
		{
			name:   "success",
			userID: 1,
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "username", "avatar_file_id", "created_at", "updated_at"}).
					AddRow(1, "test@example.com", "hash", "testuser", nil, time.Now(), nil)
				mock.ExpectQuery(`SELECT id, email, password_hash, username, avatar_file_id, created_at, updated_at`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedUser: &models.User{
				ID:       1,
				Email:    "test@example.com",
				Username: "testuser",
			},
			expectedError: nil,
		},
		{
			name:   "user not found",
			userID: 999,
			setupMocks: func() {
				mock.ExpectQuery(`SELECT id, email, password_hash, username, avatar_file_id, created_at, updated_at`).
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			expectedUser:  nil,
			expectedError: namederrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			user, err := repo.GetUserByID(ctx, tt.userID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser.ID, user.ID)
				assert.Equal(t, tt.expectedUser.Email, user.Email)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_UpdateProfile(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(store)
	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)
	ctx = middleware.WithUserID(ctx, 1)

	tests := []struct {
		name          string
		username      *string
		password      *string
		avatarFileID  *uint64
		setupMocks    func()
		expectedUser  *models.User
		expectedError error
	}{
		{
			name:     "update username",
			username: stringPtr("newusername"),
			setupMocks: func() {
				rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "username", "avatar_file_id", "created_at", "updated_at"}).
					AddRow(1, "test@example.com", "hash", "newusername", nil, time.Now(), time.Now())
				mock.ExpectQuery(`UPDATE "user"`).
					WithArgs("newusername", sqlmock.AnyArg(), 1).
					WillReturnRows(rows)
			},
			expectedUser: &models.User{
				ID:       1,
				Username: "newusername",
			},
			expectedError: nil,
		},
		{
			name:     "user not found",
			username: stringPtr("newusername"),
			setupMocks: func() {
				mock.ExpectQuery(`UPDATE "user"`).
					WithArgs("newusername", sqlmock.AnyArg(), 1).
					WillReturnError(sql.ErrNoRows)
			},
			expectedUser:  nil,
			expectedError: namederrors.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			user, err := repo.UpdateProfile(ctx, tt.username, tt.password, tt.avatarFileID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_UpdateProfile_WithPassword(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(store)
	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)
	ctx = middleware.WithUserID(ctx, 1)

	password := "newpassword"
	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "username", "avatar_file_id", "created_at", "updated_at"}).
		AddRow(1, "test@example.com", "hash", "testuser", nil, time.Now(), time.Now())

	mock.ExpectQuery(`UPDATE "user"`).
		WithArgs("newusername", password, sqlmock.AnyArg(), 1).
		WillReturnRows(rows)

	user, err := repo.UpdateProfile(ctx, stringPtr("newusername"), &password, nil)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_UpdateProfile_WithAvatarFileID(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(store)
	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)
	ctx = middleware.WithUserID(ctx, 1)

	avatarFileID := uint64(10)
	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "username", "avatar_file_id", "created_at", "updated_at"}).
		AddRow(1, "test@example.com", "hash", "testuser", 10, time.Now(), time.Now())

	mock.ExpectQuery(`UPDATE "user"`).
		WithArgs("newusername", uint64(10), sqlmock.AnyArg(), 1).
		WillReturnRows(rows)

	user, err := repo.UpdateProfile(ctx, stringPtr("newusername"), nil, &avatarFileID)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetProfile_NoUserID(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(store)
	ctx := context.Background()
	log := zerolog.Nop()
	ctx = logger.ToContext(ctx, log)
	// No user ID in context

	user, err := repo.GetProfile(ctx)
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "user not authenticated")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func stringPtr(s string) *string {
	return &s
}
