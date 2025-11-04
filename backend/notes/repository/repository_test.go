package repository

import (
	namederrors "backend/named_errors"
	"backend/store"
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

func TestNotesRepository_GetNotes(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewNotesRepository(store)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "is_archived", "is_shared", "created_at", "updated_at"}).
		AddRow(1, 1, nil, "Note 1", nil, false, false, time.Now(), time.Now()).
		AddRow(2, 1, nil, "Note 2", nil, false, false, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT id, owner_id, parent_note_id, title, icon_file_id`).
		WithArgs(1).
		WillReturnRows(rows)

	notes, err := repo.GetNotes(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(notes))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotesRepository_CreateNote(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewNotesRepository(store)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "is_archived", "is_shared", "created_at", "updated_at", "deleted_at"}).
		AddRow(1, 1, nil, "New Note", nil, false, false, time.Now(), time.Now(), nil)

	mock.ExpectQuery(`INSERT INTO note`).
		WithArgs(1, "New Note", false, false).
		WillReturnRows(rows)

	note, err := repo.CreateNote(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), note.ID)
	assert.Equal(t, "New Note", note.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotesRepository_GetNoteById(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewNotesRepository(store)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "is_archived", "is_shared", "created_at", "updated_at", "deleted_at"}).
		AddRow(1, 1, nil, "Test Note", nil, false, false, time.Now(), time.Now(), nil)

	mock.ExpectQuery(`SELECT id, owner_id, parent_note_id, title, icon_file_id`).
		WithArgs(1).
		WillReturnRows(rows)

	note, err := repo.GetNoteById(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), note.ID)
	assert.Equal(t, "Test Note", note.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotesRepository_GetNoteById_NotFound(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewNotesRepository(store)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT id, owner_id, parent_note_id, title, icon_file_id`).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	note, err := repo.GetNoteById(ctx, 999)
	assert.Error(t, err)
	assert.Equal(t, namederrors.ErrNotFound, err)
	assert.Nil(t, note)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotesRepository_UpdateNote(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewNotesRepository(store)
	ctx := context.Background()

	title := "Updated Title"
	isArchived := true

	mock.ExpectQuery(`SELECT 1 FROM note`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "is_archived", "is_shared", "created_at", "updated_at"}).
		AddRow(1, 1, nil, "Updated Title", nil, true, false, time.Now(), time.Now())

	mock.ExpectQuery(`UPDATE note SET updated_at`).
		WithArgs(sqlmock.AnyArg(), "Updated Title", true, 1).
		WillReturnRows(rows)

	note, err := repo.UpdateNote(ctx, 1, &title, &isArchived)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", note.Title)
	assert.True(t, note.IsArchived)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotesRepository_DeleteNote(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewNotesRepository(store)
	ctx := context.Background()

	mock.ExpectExec(`DELETE FROM note`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteNote(ctx, 1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotesRepository_DeleteNote_NotFound(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewNotesRepository(store)
	ctx := context.Background()

	mock.ExpectExec(`DELETE FROM note`).
		WithArgs(999).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.DeleteNote(ctx, 999)
	assert.Error(t, err)
	assert.Equal(t, namederrors.ErrNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
