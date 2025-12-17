package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"backend/notes_service/internal/constants"
	"backend/pkg/store"

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

type errorResult struct {
	err error
}

func (e *errorResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (e *errorResult) RowsAffected() (int64, error) {
	return 0, e.err
}

func TestNotesRepository_CreateNote(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

	repo := newTestRepo(db)
	ctx := context.Background()
	userID := uint64(1)
	now := time.Now().UTC()

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "header_file_id", "is_archived", "is_shared", "share_uuid", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, userID, nil, "Новая заметка", uint64(1), uint64(35), false, false, "uuid", now, now, nil)

		mock.ExpectQuery(`INSERT INTO note`).
			WithArgs(userID, nil, "Новая заметка", uint64(1), uint64(35), false, false, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
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

	t.Run("WithParentNoteID", func(t *testing.T) {
		mock.ExpectBegin()
		parentID := uint64(5)

		rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "header_file_id", "is_archived", "is_shared", "share_uuid", "created_at", "updated_at", "deleted_at"}).
			AddRow(2, userID, parentID, "Новая заметка", uint64(1), uint64(35), false, false, "uuid", now, now, nil)

		mock.ExpectQuery(`INSERT INTO note`).
			WithArgs(userID, parentID, "Новая заметка", uint64(1), uint64(35), false, false, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(rows)

		blockRows := sqlmock.NewRows([]string{"id"}).AddRow(101)
		mock.ExpectQuery(`INSERT INTO block`).
			WithArgs(2, userID, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(blockRows)

		mock.ExpectExec(`INSERT INTO block_text`).
			WithArgs(101, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		note, err := repo.CreateNote(ctx, userID, &parentID)
		assert.NoError(t, err)
		assert.NotNil(t, note)
		assert.Equal(t, uint64(2), note.ID)
		assert.NotNil(t, note.ParentNoteID)
		assert.Equal(t, parentID, *note.ParentNoteID)
	})

	t.Run("NoteInsertError", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`INSERT INTO note`).
			WithArgs(userID, nil, "Новая заметка", uint64(1), uint64(35), false, false, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New("insert error"))

		_, err := repo.CreateNote(ctx, userID, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create note")
	})

	t.Run("BlockInsertError", func(t *testing.T) {
		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "header_file_id", "is_archived", "is_shared", "share_uuid", "created_at", "updated_at", "deleted_at"}).
			AddRow(3, userID, nil, "Новая заметка", uint64(1), uint64(35), false, false, "uuid", now, now, nil)

		mock.ExpectQuery(`INSERT INTO note`).
			WithArgs(userID, nil, "Новая заметка", uint64(1), uint64(35), false, false, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(rows)

		mock.ExpectQuery(`INSERT INTO block`).
			WithArgs(3, userID, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New("block insert error"))

		_, err := repo.CreateNote(ctx, userID, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create initial block")
	})

	t.Run("BlockTextInsertError", func(t *testing.T) {
		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "header_file_id", "is_archived", "is_shared", "share_uuid", "created_at", "updated_at", "deleted_at"}).
			AddRow(4, userID, nil, "Новая заметка", uint64(1), uint64(35), false, false, "uuid", now, now, nil)

		mock.ExpectQuery(`INSERT INTO note`).
			WithArgs(userID, nil, "Новая заметка", uint64(1), uint64(35), false, false, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(rows)

		blockRows := sqlmock.NewRows([]string{"id"}).AddRow(102)
		mock.ExpectQuery(`INSERT INTO block`).
			WithArgs(4, userID, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(blockRows)

		mock.ExpectExec(`INSERT INTO block_text`).
			WithArgs(102, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New("block_text insert error"))

		_, err := repo.CreateNote(ctx, userID, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create initial block_text")
	})

	t.Run("CommitError", func(t *testing.T) {
		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "header_file_id", "is_archived", "is_shared", "share_uuid", "created_at", "updated_at", "deleted_at"}).
			AddRow(5, userID, nil, "Новая заметка", uint64(1), uint64(35), false, false, "uuid", now, now, nil)

		mock.ExpectQuery(`INSERT INTO note`).
			WithArgs(userID, nil, "Новая заметка", uint64(1), uint64(35), false, false, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(rows)

		blockRows := sqlmock.NewRows([]string{"id"}).AddRow(103)
		mock.ExpectQuery(`INSERT INTO block`).
			WithArgs(5, userID, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(blockRows)

		mock.ExpectExec(`INSERT INTO block_text`).
			WithArgs(103, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit().WillReturnError(errors.New("commit error"))

		_, err := repo.CreateNote(ctx, userID, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to commit create note transaction")
	})
}

func TestNotesRepository_GetNoteById(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

	repo := newTestRepo(db)
	ctx := context.Background()
	noteID := uint64(1)
	userID := uint64(1)

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "header_file_id", "is_archived", "is_shared", "share_uuid", "created_at", "updated_at", "deleted_at", "is_favorite"}).
			AddRow(noteID, userID, nil, "Test Note", nil, nil, false, false, "uuid", time.Now(), time.Now(), nil, false)

		mock.ExpectQuery(`SELECT (.+) FROM note n`).
			WithArgs(noteID, userID).
			WillReturnRows(rows)

		note, err := repo.GetNoteById(ctx, noteID, userID)
		assert.NoError(t, err)
		assert.NotNil(t, note)
		assert.Equal(t, noteID, note.ID)
		assert.Nil(t, note.DeletedAt)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(`SELECT (.+) FROM note n`).
			WithArgs(noteID, userID).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetNoteById(ctx, noteID, userID)
		assert.ErrorIs(t, err, constants.ErrNotFound)
	})

	t.Run("WithNullableFields", func(t *testing.T) {
		parentID := uint64(10)
		iconID := uint64(5)
		headerID := uint64(20)
		shareUUID := "test-uuid"

		rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "header_file_id", "is_archived", "is_shared", "share_uuid", "created_at", "updated_at", "deleted_at", "is_favorite"}).
			AddRow(noteID, userID, parentID, "Test Note", iconID, headerID, false, true, shareUUID, time.Now(), time.Now(), nil, true)

		mock.ExpectQuery(`SELECT (.+) FROM note n`).
			WithArgs(noteID, userID).
			WillReturnRows(rows)

		note, err := repo.GetNoteById(ctx, noteID, userID)
		assert.NoError(t, err)
		assert.NotNil(t, note)
		assert.NotNil(t, note.ParentNoteID)
		assert.Equal(t, parentID, *note.ParentNoteID)
		assert.NotNil(t, note.IconFileID)
		assert.Equal(t, iconID, *note.IconFileID)
		assert.NotNil(t, note.HeaderFileID)
		assert.Equal(t, headerID, *note.HeaderFileID)
		assert.NotNil(t, note.ShareUUID)
		assert.Equal(t, shareUUID, *note.ShareUUID)
		assert.True(t, note.IsFavorite)
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mock.ExpectQuery(`SELECT (.+) FROM note n`).
			WithArgs(noteID, userID).
			WillReturnError(errors.New("database error"))

		_, err := repo.GetNoteById(ctx, noteID, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get note")
	})
}

func TestNotesRepository_UpdateNote(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

	repo := newTestRepo(db)
	ctx := context.Background()
	noteID := uint64(1)
	title := "Updated Title"
	isArchived := true

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(`SELECT 1 FROM note`).WithArgs(noteID).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(1))

		rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "header_file_id", "is_archived", "is_shared", "share_uuid", "created_at", "updated_at"}).
			AddRow(noteID, 1, nil, title, nil, nil, isArchived, false, "uuid", time.Now(), time.Now())

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

	t.Run("OnlyTitle", func(t *testing.T) {
		mock.ExpectQuery(`SELECT 1 FROM note`).WithArgs(noteID).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(1))

		rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "header_file_id", "is_archived", "is_shared", "share_uuid", "created_at", "updated_at"}).
			AddRow(noteID, 1, nil, title, nil, nil, false, false, "uuid", time.Now(), time.Now())

		mock.ExpectQuery(`UPDATE note`).
			WithArgs(sqlmock.AnyArg(), title, noteID).
			WillReturnRows(rows)

		note, err := repo.UpdateNote(ctx, noteID, &title, nil)
		assert.NoError(t, err)
		assert.Equal(t, title, note.Title)
	})

	t.Run("OnlyIsArchived", func(t *testing.T) {
		mock.ExpectQuery(`SELECT 1 FROM note`).WithArgs(noteID).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(1))

		rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "header_file_id", "is_archived", "is_shared", "share_uuid", "created_at", "updated_at"}).
			AddRow(noteID, 1, nil, "Original Title", nil, nil, isArchived, false, "uuid", time.Now(), time.Now())

		mock.ExpectQuery(`UPDATE note`).
			WithArgs(sqlmock.AnyArg(), isArchived, noteID).
			WillReturnRows(rows)

		note, err := repo.UpdateNote(ctx, noteID, nil, &isArchived)
		assert.NoError(t, err)
		assert.Equal(t, isArchived, note.IsArchived)
	})

	t.Run("CheckExistenceError", func(t *testing.T) {
		mock.ExpectQuery(`SELECT 1 FROM note`).WithArgs(noteID).WillReturnError(errors.New("check error"))

		_, err := repo.UpdateNote(ctx, noteID, &title, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to check note existence")
	})
}

func TestNotesRepository_DeleteNote(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

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

	t.Run("DatabaseError", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM note`).WithArgs(noteID).WillReturnError(errors.New("database error"))
		err := repo.DeleteNote(ctx, noteID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete note")
	})

	t.Run("RowsAffectedError", func(t *testing.T) {
		result := &errorResult{err: errors.New("rows affected error")}
		mock.ExpectExec(`DELETE FROM note`).WithArgs(noteID).WillReturnResult(result)
		err := repo.DeleteNote(ctx, noteID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get rows affected")
	})
}

func TestNotesRepository_GetNotes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

	repo := newTestRepo(db)
	ctx := context.Background()
	userID := uint64(1)

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "header_file_id", "is_archived", "is_shared", "share_uuid", "created_at", "updated_at", "is_favorite"}).
			AddRow(1, userID, nil, "Note 1", nil, nil, false, false, "uuid1", time.Now(), time.Now(), false).
			AddRow(2, userID, nil, "Note 2", nil, nil, false, true, "uuid2", time.Now(), time.Now(), true)

		mock.ExpectQuery(`SELECT DISTINCT`).WithArgs(userID).WillReturnRows(rows)

		notes, err := repo.GetNotes(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, notes, 2)
		assert.Equal(t, "Note 1", notes[0].Title)
		assert.Equal(t, "Note 2", notes[1].Title)
		assert.True(t, notes[1].IsFavorite)
	})

	t.Run("EmptyResults", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "header_file_id", "is_archived", "is_shared", "share_uuid", "created_at", "updated_at", "is_favorite"})

		mock.ExpectQuery(`SELECT DISTINCT`).WithArgs(userID).WillReturnRows(rows)

		notes, err := repo.GetNotes(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, notes, 0)
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mock.ExpectQuery(`SELECT DISTINCT`).WithArgs(userID).WillReturnError(errors.New("database error"))

		_, err := repo.GetNotes(ctx, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list notes")
	})

	t.Run("ScanError", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "header_file_id", "is_archived", "is_shared", "share_uuid", "created_at", "updated_at", "is_favorite"}).
			AddRow("invalid", userID, nil, "Note", nil, nil, false, false, "uuid", time.Now(), time.Now(), false)

		mock.ExpectQuery(`SELECT DISTINCT`).WithArgs(userID).WillReturnRows(rows)

		_, err := repo.GetNotes(ctx, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to scan note")
	})

	t.Run("RowsError", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "header_file_id", "is_archived", "is_shared", "share_uuid", "created_at", "updated_at", "is_favorite"}).
			AddRow(1, userID, nil, "Note", nil, nil, false, false, "uuid", time.Now(), time.Now(), false).
			RowError(0, errors.New("row error"))

		mock.ExpectQuery(`SELECT DISTINCT`).WithArgs(userID).WillReturnRows(rows)

		_, err := repo.GetNotes(ctx, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error iterating notes")
	})

	t.Run("WithNullableFields", func(t *testing.T) {
		parentID := uint64(5)
		iconID := uint64(10)
		headerID := uint64(15)
		shareUUID := "test-uuid"

		rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "header_file_id", "is_archived", "is_shared", "share_uuid", "created_at", "updated_at", "is_favorite"}).
			AddRow(3, userID, parentID, "Note with fields", iconID, headerID, false, true, shareUUID, time.Now(), time.Now(), false)

		mock.ExpectQuery(`SELECT DISTINCT`).WithArgs(userID).WillReturnRows(rows)

		notes, err := repo.GetNotes(ctx, userID)
		assert.NoError(t, err)
		assert.Len(t, notes, 1)
		assert.NotNil(t, notes[0].ParentNoteID)
		assert.Equal(t, parentID, *notes[0].ParentNoteID)
		assert.NotNil(t, notes[0].IconFileID)
		assert.Equal(t, iconID, *notes[0].IconFileID)
		assert.NotNil(t, notes[0].HeaderFileID)
		assert.Equal(t, headerID, *notes[0].HeaderFileID)
		assert.NotNil(t, notes[0].ShareUUID)
		assert.Equal(t, shareUUID, *notes[0].ShareUUID)
	})
}

func TestNotesRepository_AddFavorite(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

	repo := newTestRepo(db)
	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(`INSERT INTO favorite`).WithArgs(userID, noteID, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
		err := repo.AddFavorite(ctx, userID, noteID)
		assert.NoError(t, err)
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mock.ExpectExec(`INSERT INTO favorite`).WithArgs(userID, noteID, sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnError(errors.New("database error"))
		err := repo.AddFavorite(ctx, userID, noteID)
		assert.Error(t, err)
	})
}

func TestNotesRepository_RemoveFavorite(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

	repo := newTestRepo(db)
	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM favorite`).WithArgs(userID, noteID).WillReturnResult(sqlmock.NewResult(0, 1))
		err := repo.RemoveFavorite(ctx, userID, noteID)
		assert.NoError(t, err)
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM favorite`).WithArgs(userID, noteID).WillReturnError(errors.New("database error"))
		err := repo.RemoveFavorite(ctx, userID, noteID)
		assert.Error(t, err)
	})
}

func TestNotesRepository_CheckNoteOwnership(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

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
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

	repo := newTestRepo(db)
	ctx := context.Background()
	shareUUID := "share-uuid"

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "header_file_id", "is_archived", "is_shared", "share_uuid", "public_access_level", "created_at", "updated_at", "deleted_at"}).
			AddRow(1, 1, nil, "Note", nil, nil, false, true, shareUUID, "viewer", time.Now(), time.Now(), nil)

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

	t.Run("DatabaseError", func(t *testing.T) {
		mock.ExpectQuery(`SELECT (.+) FROM note`).WithArgs(shareUUID).WillReturnError(errors.New("database error"))
		_, err := repo.GetNoteByShareUUID(ctx, shareUUID)
		assert.Error(t, err)
	})

	t.Run("WithNullableFields", func(t *testing.T) {
		parentID := uint64(10)
		iconID := uint64(5)
		headerID := uint64(20)
		publicAccessLevel := "editor"

		rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "header_file_id", "is_archived", "is_shared", "share_uuid", "public_access_level", "created_at", "updated_at", "deleted_at"}).
			AddRow(2, 1, parentID, "Shared Note", iconID, headerID, false, true, shareUUID, publicAccessLevel, time.Now(), time.Now(), nil)

		mock.ExpectQuery(`SELECT (.+) FROM note`).WithArgs(shareUUID).WillReturnRows(rows)

		note, err := repo.GetNoteByShareUUID(ctx, shareUUID)
		assert.NoError(t, err)
		assert.NotNil(t, note)
		assert.NotNil(t, note.ParentNoteID)
		assert.Equal(t, parentID, *note.ParentNoteID)
		assert.NotNil(t, note.IconFileID)
		assert.Equal(t, iconID, *note.IconFileID)
		assert.NotNil(t, note.HeaderFileID)
		assert.Equal(t, headerID, *note.HeaderFileID)
	})
}

func TestNotesRepository_SetIcon(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

	repo := newTestRepo(db)
	ctx := context.Background()
	noteID := uint64(1)
	iconFileID := uint64(10)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(`UPDATE note SET icon_file_id`).
			WithArgs(iconFileID, sqlmock.AnyArg(), noteID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.SetIcon(ctx, noteID, iconFileID)
		assert.NoError(t, err)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectExec(`UPDATE note SET icon_file_id`).
			WithArgs(iconFileID, sqlmock.AnyArg(), noteID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.SetIcon(ctx, noteID, iconFileID)
		assert.ErrorIs(t, err, constants.ErrNotFound)
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mock.ExpectExec(`UPDATE note SET icon_file_id`).
			WithArgs(iconFileID, sqlmock.AnyArg(), noteID).
			WillReturnError(errors.New("database error"))

		err := repo.SetIcon(ctx, noteID, iconFileID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to set icon")
	})

	t.Run("RowsAffectedError", func(t *testing.T) {
		result := &errorResult{err: errors.New("rows affected error")}
		mock.ExpectExec(`UPDATE note SET icon_file_id`).
			WithArgs(iconFileID, sqlmock.AnyArg(), noteID).
			WillReturnResult(result)

		err := repo.SetIcon(ctx, noteID, iconFileID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get rows affected")
	})
}

func TestNotesRepository_SetHeader(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

	repo := newTestRepo(db)
	ctx := context.Background()
	noteID := uint64(1)
	headerFileID := uint64(20)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(`UPDATE note SET header_file_id`).
			WithArgs(headerFileID, sqlmock.AnyArg(), noteID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.SetHeader(ctx, noteID, headerFileID)
		assert.NoError(t, err)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectExec(`UPDATE note SET header_file_id`).
			WithArgs(headerFileID, sqlmock.AnyArg(), noteID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.SetHeader(ctx, noteID, headerFileID)
		assert.ErrorIs(t, err, constants.ErrNotFound)
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mock.ExpectExec(`UPDATE note SET header_file_id`).
			WithArgs(headerFileID, sqlmock.AnyArg(), noteID).
			WillReturnError(errors.New("database error"))

		err := repo.SetHeader(ctx, noteID, headerFileID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to set header")
	})

	t.Run("RowsAffectedError", func(t *testing.T) {
		result := &errorResult{err: errors.New("rows affected error")}
		mock.ExpectExec(`UPDATE note SET header_file_id`).
			WithArgs(headerFileID, sqlmock.AnyArg(), noteID).
			WillReturnResult(result)

		err := repo.SetHeader(ctx, noteID, headerFileID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get rows affected")
	})
}

func TestNotesRepository_SearchNotes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

	repo := newTestRepo(db)
	ctx := context.Background()
	userID := uint64(1)
	searchQuery := "test query"

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"note_id", "title", "highlighted_title", "content_snippet", "rank", "updated_at"}).
			AddRow(uint64(1), "Test Note 1", "<mark>Test</mark> Note 1", "content snippet 1", 0.5, time.Now()).
			AddRow(uint64(2), "Test Note 2", "<mark>Test</mark> Note 2", "content snippet 2", 0.3, time.Now())

		mock.ExpectQuery(`SELECT`).
			WithArgs(searchQuery, userID).
			WillReturnRows(rows)

		result, err := repo.SearchNotes(ctx, userID, searchQuery)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, result.Count)
		assert.Len(t, result.Results, 2)
		assert.Equal(t, uint64(1), result.Results[0].NoteID)
		assert.Equal(t, "Test Note 1", result.Results[0].Title)
		assert.Equal(t, "<mark>Test</mark> Note 1", result.Results[0].HighlightedTitle)
		assert.Equal(t, "content snippet 1", result.Results[0].ContentSnippet)
		assert.Equal(t, float32(0.5), result.Results[0].Rank)
	})

	t.Run("EmptyResults", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"note_id", "title", "highlighted_title", "content_snippet", "rank", "updated_at"})

		mock.ExpectQuery(`SELECT`).
			WithArgs(searchQuery, userID).
			WillReturnRows(rows)

		result, err := repo.SearchNotes(ctx, userID, searchQuery)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 0, result.Count)
		assert.Len(t, result.Results, 0)
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mock.ExpectQuery(`SELECT`).
			WithArgs(searchQuery, userID).
			WillReturnError(errors.New("database error"))

		result, err := repo.SearchNotes(ctx, userID, searchQuery)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to execute search query")
	})

	t.Run("ScanError", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"note_id", "title", "highlighted_title", "content_snippet", "rank", "updated_at"}).
			AddRow("invalid", "Test Note", "Highlighted", "Snippet", 0.5, time.Now())

		mock.ExpectQuery(`SELECT`).
			WithArgs(searchQuery, userID).
			WillReturnRows(rows)

		result, err := repo.SearchNotes(ctx, userID, searchQuery)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to scan search result")
	})

	t.Run("RowsError", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"note_id", "title", "highlighted_title", "content_snippet", "rank", "updated_at"}).
			AddRow(uint64(1), "Test Note", "Highlighted", "Snippet", 0.5, time.Now()).
			RowError(0, errors.New("row error"))

		mock.ExpectQuery(`SELECT`).
			WithArgs(searchQuery, userID).
			WillReturnRows(rows)

		result, err := repo.SearchNotes(ctx, userID, searchQuery)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "error iterating search results")
	})
}
