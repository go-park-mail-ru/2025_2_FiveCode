package repository

import (
	"backend/notes_service/internal/constants"
	"backend/pkg/store"
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
	return d.DB.BeginTx(ctx, opts)
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

func newTestRepo(db *sql.DB) *NotesRepository {
	return NewNotesRepository(&testStoreDB{DB: db})
}

func TestNotesRepository_CreateNote(t *testing.T) {
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
		mock.ExpectBegin()
		
		rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "is_archived", "is_shared", "share_uuid", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, userID, nil, "Новая заметка", nil, false, false, "uuid", now, now, nil)

		mock.ExpectQuery(`INSERT INTO note`).
			WithArgs(userID, nil, "Новая заметка", false, false, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(rows)

		blockRows := sqlmock.NewRows([]string{"id"}).AddRow(100)
		mock.ExpectQuery(`INSERT INTO block`).
			WithArgs(1, userID, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(blockRows)

		mock.ExpectExec(`INSERT INTO block_text`).
			WithArgs(100, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		note, err := repo.CreateNote(ctx, userID, nil)
		assert.NoError(t, err)
		assert.NotNil(t, note)
		assert.Equal(t, uint64(1), note.ID)
	})

	t.Run("BeginTxError", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(errors.New("tx error"))
		_, err := repo.CreateNote(ctx, userID, nil)
		assert.Error(t, err)
	})
}

func TestNotesRepository_GetNoteById(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	noteID := uint64(1)
	userID := uint64(1)

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "is_archived", "is_shared", "share_uuid", "created_at", "updated_at", "deleted_at", "is_favorite"}).
			AddRow(noteID, userID, nil, "Test Note", nil, false, false, "uuid", time.Now(), time.Now(), nil, false)

		mock.ExpectQuery(`SELECT (.+) FROM note n`).
			WithArgs(noteID, userID).
			WillReturnRows(rows)

		note, err := repo.GetNoteById(ctx, noteID, userID)
		assert.NoError(t, err)
		assert.NotNil(t, note)
		assert.Equal(t, noteID, note.ID)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(`SELECT (.+) FROM note n`).
			WithArgs(noteID, userID).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetNoteById(ctx, noteID, userID)
		assert.ErrorIs(t, err, constants.ErrNotFound)
	})
}

func TestNotesRepository_UpdateNote(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	noteID := uint64(1)
	title := "Updated Title"
	isArchived := true

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(`SELECT 1 FROM note`).WithArgs(noteID).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(1))

		rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "is_archived", "is_shared", "share_uuid", "created_at", "updated_at"}).
			AddRow(noteID, 1, nil, title, nil, isArchived, false, "uuid", time.Now(), time.Now())

		mock.ExpectQuery(`UPDATE note`).
			WithArgs(sqlmock.AnyArg(), title, isArchived, noteID).
			WillReturnRows(rows)

		note, err := repo.UpdateNote(ctx, noteID, &title, &isArchived)
		assert.NoError(t, err)
		assert.Equal(t, title, note.Title)
		assert.Equal(t, isArchived, note.IsArchived)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(`SELECT 1 FROM note`).WithArgs(noteID).WillReturnError(sql.ErrNoRows)
		_, err := repo.UpdateNote(ctx, noteID, &title, nil)
		assert.ErrorIs(t, err, constants.ErrNotFound)
	})
}

func TestNotesRepository_DeleteNote(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	noteID := uint64(1)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM note`).WithArgs(noteID).WillReturnResult(sqlmock.NewResult(0, 1))
		err := repo.DeleteNote(ctx, noteID)
		assert.NoError(t, err)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM note`).WithArgs(noteID).WillReturnResult(sqlmock.NewResult(0, 0))
		err := repo.DeleteNote(ctx, noteID)
		assert.ErrorIs(t, err, constants.ErrNotFound)
	})
}

func TestNotesRepository_GetNotes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	userID := uint64(1)

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "is_archived", "is_shared", "share_uuid", "created_at", "updated_at", "is_favorite"}).
			AddRow(1, userID, nil, "Note 1", nil, false, false, "uuid1", time.Now(), time.Now(), false).
			AddRow(2, userID, nil, "Note 2", nil, false, true, "uuid2", time.Now(), time.Now(), true)

		mock.ExpectQuery(`SELECT DISTINCT`).WithArgs(userID).WillReturnRows(rows)

		notes, err := repo.GetNotes(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, notes, 2)
		assert.Equal(t, "Note 1", notes[0].Title)
		assert.Equal(t, "Note 2", notes[1].Title)
		assert.True(t, notes[1].IsFavorite)
	})
}

func TestNotesRepository_AddFavorite(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(`INSERT INTO favorite`).WithArgs(userID, noteID, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
		err := repo.AddFavorite(ctx, userID, noteID)
		assert.NoError(t, err)
	})
}

func TestNotesRepository_RemoveFavorite(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM favorite`).WithArgs(userID, noteID).WillReturnResult(sqlmock.NewResult(0, 1))
		err := repo.RemoveFavorite(ctx, userID, noteID)
		assert.NoError(t, err)
	})
}

func TestNotesRepository_CheckNoteOwnership(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)

	t.Run("IsOwner", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"owner_id"}).AddRow(userID)
		mock.ExpectQuery(`SELECT owner_id FROM note`).WithArgs(noteID).WillReturnRows(rows)
		isOwner, err := repo.CheckNoteOwnership(ctx, noteID, userID)
		assert.NoError(t, err)
		assert.True(t, isOwner)
	})

	t.Run("NotOwner", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"owner_id"}).AddRow(userID + 1)
		mock.ExpectQuery(`SELECT owner_id FROM note`).WithArgs(noteID).WillReturnRows(rows)
		isOwner, err := repo.CheckNoteOwnership(ctx, noteID, userID)
		assert.NoError(t, err)
		assert.False(t, isOwner)
	})
	
	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(`SELECT owner_id FROM note`).WithArgs(noteID).WillReturnError(sql.ErrNoRows)
		_, err := repo.CheckNoteOwnership(ctx, noteID, userID)
		assert.ErrorIs(t, err, constants.ErrNotFound)
	})
}

func TestNotesRepository_GetNoteByShareUUID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	shareUUID := "share-uuid"

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "is_archived", "is_shared", "share_uuid", "public_access_level", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, 1, nil, "Note", nil, false, true, shareUUID, "viewer", time.Now(), time.Now(), nil)
		
		mock.ExpectQuery(`SELECT (.+) FROM note`).WithArgs(shareUUID).WillReturnRows(rows)
		
		note, err := repo.GetNoteByShareUUID(ctx, shareUUID)
		assert.NoError(t, err)
		assert.Equal(t, shareUUID, *note.ShareUUID)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(`SELECT (.+) FROM note`).WithArgs(shareUUID).WillReturnError(sql.ErrNoRows)
		_, err := repo.GetNoteByShareUUID(ctx, shareUUID)
		assert.ErrorIs(t, err, constants.ErrNotFound)
	})
}
