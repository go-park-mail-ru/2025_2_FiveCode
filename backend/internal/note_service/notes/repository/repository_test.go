package repository

import (
	namederrors "backend/named_errors"
	store2 "backend/pkg/store"
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *store2.Store) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}

	s := &store2.Store{
		Postgres: &store2.PostgresDB{DB: db},
	}

	return db, mock, s
}

func TestNotesRepository_GetNoteById(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewNotesRepository(store)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "is_archived", "is_shared", "created_at", "updated_at", "deleted_at", "is_favorite"}).
		AddRow(1, 1, nil, "Test Note", nil, false, false, time.Now(), time.Now(), nil, false)

	mock.ExpectQuery(`SELECT n.id, n.owner_id, n.parent_note_id, n.title`).
		WithArgs(uint64(1), uint64(1)).
		WillReturnRows(rows)

	note, err := repo.GetNoteById(ctx, 1, 1)
	assert.NoError(t, err)
	assert.NotNil(t, note)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotesRepository_GetNoteById_NotFound(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewNotesRepository(store)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT n.id, n.owner_id, n.parent_note_id, n.title`).
		WithArgs(uint64(999), uint64(1)).
		WillReturnError(sql.ErrNoRows)

	note, err := repo.GetNoteById(ctx, 999, 1)
	assert.Error(t, err)
	assert.Equal(t, namederrors.ErrNotFound, err)
	assert.Nil(t, note)
	assert.NoError(t, mock.ExpectationsWereMet())
}
