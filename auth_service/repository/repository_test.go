package repository

import (
	"backend/pkg/store"
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
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

func TestAuthRepository_GetUserByEmail(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewAuthRepository(store)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "username", "avatar_file_id", "created_at", "updated_at"}).
		AddRow(1, "test@example.com", "hash", "testuser", nil, time.Now(), nil)

	mock.ExpectQuery(`SELECT id, email, password_hash, username, avatar_file_id, created_at, updated_at`).
		WithArgs("test@example.com").
		WillReturnRows(rows)

	user, err := repo.GetUserByEmail(ctx, "test@example.com")
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_GetUserByEmail_NotFound(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewAuthRepository(store)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT id, email, password_hash, username, avatar_file_id, created_at, updated_at`).
		WithArgs("notfound@example.com").
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetUserByEmail(ctx, "notfound@example.com")
	assert.Error(t, err)
	assert.Equal(t, namederrors.ErrNotFound, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_CreateUser(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewAuthRepository(store)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "username", "avatar_file_id", "created_at", "updated_at"}).
		AddRow(1, "newuser@example.com", "hash", "newuser", nil, time.Now(), nil)

	mock.ExpectQuery(`INSERT INTO "user"`).
		WithArgs("newuser@example.com", "hash", "newuser").
		WillReturnRows(rows)

	user, err := repo.CreateUser(ctx, "newuser@example.com", "hash")
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), user.ID)
	assert.Equal(t, "newuser@example.com", user.Email)
	assert.Equal(t, "newuser", user.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_CreateUser_AlreadyExists(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewAuthRepository(store)
	ctx := context.Background()

	// "duplicate key"
	mock.ExpectQuery(`INSERT INTO "user"`).
		WithArgs("existing@example.com", "hash", "existing").
		WillReturnError(sql.ErrConnDone)

	user, err := repo.CreateUser(ctx, "existing@example.com", "hash")
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_GetUserByID(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewAuthRepository(store)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "username", "avatar_file_id", "created_at", "updated_at"}).
		AddRow(1, "test@example.com", "hash", "testuser", nil, time.Now(), nil)

	mock.ExpectQuery(`SELECT id, email, password_hash, username, avatar_file_id, created_at, updated_at`).
		WithArgs(1).
		WillReturnRows(rows)

	user, err := repo.GetUserByID(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthRepository_GetUserByID_NotFound(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewAuthRepository(store)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT id, email, password_hash, username, avatar_file_id, created_at, updated_at`).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetUserByID(ctx, 999)
	assert.Error(t, err)
	assert.Equal(t, namederrors.ErrNotFound, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}
