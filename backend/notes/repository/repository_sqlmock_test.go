package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestNotesRepository_CreateNote_Success(t *testing.T) {
    db, mock, s := setupTestDB(t)
    defer db.Close()

    repo := NewNotesRepository(s)
    ctx := context.Background()

    // Expect begin
    mock.ExpectBegin()

    // Return created note row
    noteRows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "is_archived", "is_shared", "created_at", "updated_at", "deleted_at"}).
        AddRow(1, 1, nil, "New Note", nil, false, false, time.Now(), time.Now(), nil)
    mock.ExpectQuery("INSERT INTO note").
        WithArgs(uint64(1), sqlmock.AnyArg(), false, false).
        WillReturnRows(noteRows)

    // Expect block insert
    blockRows := sqlmock.NewRows([]string{"id"}).AddRow(2)
    mock.ExpectQuery("INSERT INTO block").WithArgs(uint64(1), sqlmock.AnyArg()).WillReturnRows(blockRows)

    // Expect block_text insert
    mock.ExpectExec("INSERT INTO block_text").WithArgs(uint64(2)).WillReturnResult(sqlmock.NewResult(1, 1))

    mock.ExpectCommit()

    note, err := repo.CreateNote(ctx, 1)
    assert.NoError(t, err)
    assert.NotNil(t, note)
    assert.Equal(t, uint64(1), note.ID)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotesRepository_GetNotes_Success(t *testing.T) {
    db, mock, s := setupTestDB(t)
    defer db.Close()

    repo := NewNotesRepository(s)
    ctx := context.Background()

    rows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "is_archived", "is_shared", "created_at", "updated_at", "is_favorite"}).
        AddRow(1, 1, nil, "Test Note", nil, false, false, time.Now(), time.Now(), false)

    mock.ExpectQuery("SELECT n.id, n.owner_id, n.parent_note_id").WithArgs(uint64(1)).WillReturnRows(rows)

    notes, err := repo.GetNotes(ctx, 1)
    assert.NoError(t, err)
    if assert.Len(t, notes, 1) {
        assert.Equal(t, uint64(1), notes[0].ID)
    }
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotesRepository_UpdateNote_Success(t *testing.T) {
    db, mock, s := setupTestDB(t)
    defer db.Close()

    repo := NewNotesRepository(s)
    ctx := context.Background()

    // Check exists
    mock.ExpectQuery("SELECT 1 FROM note WHERE id =").WithArgs(uint64(1)).WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

    // Update returning row
    updatedRows := sqlmock.NewRows([]string{"id", "owner_id", "parent_note_id", "title", "icon_file_id", "is_archived", "is_shared", "created_at", "updated_at"}).
        AddRow(1, 1, nil, "Updated", nil, false, false, time.Now(), time.Now())

    mock.ExpectQuery("UPDATE note SET updated_at").WithArgs(sqlmock.AnyArg(), "Updated", uint64(1)).WillReturnRows(updatedRows)

    title := "Updated"
    note, err := repo.UpdateNote(ctx, 1, &title, nil)
    assert.NoError(t, err)
    assert.NotNil(t, note)
    assert.Equal(t, "Updated", note.Title)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotesRepository_DeleteNote_NotFound(t *testing.T) {
    db, mock, s := setupTestDB(t)
    defer db.Close()

    repo := NewNotesRepository(s)
    ctx := context.Background()

    mock.ExpectExec("DELETE FROM note WHERE id =").WithArgs(uint64(999)).WillReturnResult(sqlmock.NewResult(0, 0))

    err := repo.DeleteNote(ctx, 999)
    assert.Error(t, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotesRepository_AddRemoveFavorite(t *testing.T) {
    db, mock, s := setupTestDB(t)
    defer db.Close()

    repo := NewNotesRepository(s)
    ctx := context.Background()

    mock.ExpectExec("INSERT INTO favorite").WithArgs(uint64(1), uint64(2)).WillReturnResult(sqlmock.NewResult(1, 1))
    err := repo.AddFavorite(ctx, 1, 2)
    assert.NoError(t, err)

    mock.ExpectExec("DELETE FROM favorite WHERE user_id =").WithArgs(uint64(1), uint64(2)).WillReturnResult(sqlmock.NewResult(1, 1))
    err = repo.RemoveFavorite(ctx, 1, 2)
    assert.NoError(t, err)

    assert.NoError(t, mock.ExpectationsWereMet())
}
