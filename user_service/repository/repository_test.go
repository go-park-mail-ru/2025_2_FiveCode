package repository

import (
	"backend/pkg/store"
	"backend/user_service/internal/constants"
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

type testStoreDB struct {
	*sql.DB
}

func (d *testStoreDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (store.Tx, error) {
	return nil, errors.New("not implemented")
}

func (d *testStoreDB) GetSQLDB() *sql.DB {
	return d.DB
}

func (d *testStoreDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return d.DB.QueryRowContext(ctx, query, args...)
}

func (d *testStoreDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.DB.ExecContext(ctx, query, args...)
}

func (d *testStoreDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return d.DB.QueryContext(ctx, query, args...)
}

func (d *testStoreDB) Close() error {
	return d.DB.Close()
}

func newTestRepo(db *sql.DB) *UserRepository {
	return NewUserRepository(&testStoreDB{DB: db})
}

func TestUserRepository_CreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()

	email := "test@example.com"
	password := "hashed_password"
	username := "testuser"

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
		mock.ExpectQuery(`INSERT INTO "user"`).
			WithArgs(email, password, username, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(rows)

		id, err := repo.CreateUser(ctx, email, password, username)
		assert.NoError(t, err)
		assert.Equal(t, uint64(1), id)
	})

	t.Run("DuplicateUser", func(t *testing.T) {
		mock.ExpectQuery(`INSERT INTO "user"`).
			WithArgs(email, password, username, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New("duplicate key value violates unique constraint"))

		id, err := repo.CreateUser(ctx, email, password, username)
		assert.Error(t, err)
		assert.Equal(t, constants.ErrUserExists, err)
		assert.Equal(t, uint64(0), id)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(`INSERT INTO "user"`).
			WithArgs(email, password, username, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New("db error"))

		id, err := repo.CreateUser(ctx, email, password, username)
		assert.Error(t, err)
		assert.NotEqual(t, constants.ErrUserExists, err)
		assert.Equal(t, uint64(0), id)
	})
}

func TestUserRepository_GetUserByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	userID := uint64(1)
	now := time.Now().UTC()

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "username", "avatar_file_id", "created_at", "updated_at"}).
			AddRow(userID, "test@example.com", "hash", "testuser", nil, now, now)

		mock.ExpectQuery(`SELECT id, email, password_hash, username, avatar_file_id, created_at, updated_at FROM "user"`).
			WithArgs(userID).
			WillReturnRows(rows)

		user, err := repo.GetUserByID(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, "test@example.com", user.Email)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, email, password_hash, username, avatar_file_id, created_at, updated_at FROM "user"`).
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		user, err := repo.GetUserByID(ctx, userID)
		assert.Error(t, err)
		assert.Equal(t, constants.ErrNotFound, err)
		assert.Nil(t, user)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, email, password_hash, username, avatar_file_id, created_at, updated_at FROM "user"`).
			WithArgs(userID).
			WillReturnError(errors.New("db error"))

		user, err := repo.GetUserByID(ctx, userID)
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	email := "test@example.com"
	now := time.Now().UTC()

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "username", "avatar_file_id", "created_at", "updated_at"}).
			AddRow(1, email, "hash", "testuser", nil, now, now)

		mock.ExpectQuery(`SELECT id, email, password_hash, username, avatar_file_id, created_at, updated_at FROM "user"`).
			WithArgs(email).
			WillReturnRows(rows)

		user, err := repo.GetUserByEmail(ctx, email)
		assert.NoError(t, err)
		assert.Equal(t, uint64(1), user.ID)
		assert.Equal(t, email, user.Email)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, email, password_hash, username, avatar_file_id, created_at, updated_at FROM "user"`).
			WithArgs(email).
			WillReturnError(sql.ErrNoRows)

		user, err := repo.GetUserByEmail(ctx, email)
		assert.Error(t, err)
		assert.Equal(t, constants.ErrNotFound, err)
		assert.Nil(t, user)
	})
}

func TestUserRepository_DeleteUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	userID := uint64(1)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM "user"`).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteUser(ctx, userID)
		assert.NoError(t, err)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM "user"`).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.DeleteUser(ctx, userID)
		assert.Error(t, err)
		assert.Equal(t, constants.ErrNotFound, err)
	})
}

func TestUserRepository_UpdateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	userID := uint64(1)
	username := "newusername"
	now := time.Now().UTC()

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "username", "avatar_file_id", "created_at", "updated_at"}).
			AddRow(userID, "email", "hash", username, nil, now, now)

		mock.ExpectQuery(`UPDATE "user"`).
			WithArgs(username, sqlmock.AnyArg(), userID).
			WillReturnRows(rows)

		user, err := repo.UpdateUser(ctx, userID, &username, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, username, user.Username)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(`UPDATE "user"`).
			WithArgs(username, sqlmock.AnyArg(), userID).
			WillReturnError(sql.ErrNoRows)

		user, err := repo.UpdateUser(ctx, userID, &username, nil, nil)
		assert.Error(t, err)
		assert.Equal(t, constants.ErrNotFound, err)
		assert.Nil(t, user)
	})
}
