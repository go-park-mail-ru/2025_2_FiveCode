package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestBlocksRepository_CreateTextBlock_Success(t *testing.T) {
    db, mock, s := setupTestDB(t)
    defer db.Close()

    repo := NewBlocksRepository(s)
    ctx := context.Background()

    mock.ExpectBegin()
    mock.ExpectQuery("INSERT INTO block").WithArgs(uint64(1), sqlmock.AnyArg(), uint64(1)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
    mock.ExpectExec("INSERT INTO block_text").WithArgs(uint64(10), "").WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectCommit()

    // After commit, GetBlockByID will be called: return a row
    row := sqlmock.NewRows([]string{"id", "note_id", "type", "position", "created_at", "updated_at", "text", "file_url"}).
        AddRow(10, 1, "text", 1.0, time.Now(), time.Now(), "", nil)
    mock.ExpectQuery("SELECT b.id, b.note_id, b.type, b.position").WithArgs(uint64(10)).WillReturnRows(row)
    // When block type is text, repository will query formats; return empty result
    mock.ExpectQuery("SELECT btf.id, btf.block_text_id, btf.start_offset").WithArgs(uint64(10)).WillReturnRows(sqlmock.NewRows([]string{"id", "block_text_id", "start_offset", "end_offset", "bold", "italic", "underline", "strikethrough", "link", "font", "size"}))

    b, err := repo.CreateTextBlock(ctx, 1, 1.0, 1)
    assert.NoError(t, err)
    assert.NotNil(t, b)
    assert.Equal(t, uint64(10), b.ID)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBlocksRepository_CreateAttachmentBlock_Success(t *testing.T) {
    db, mock, s := setupTestDB(t)
    defer db.Close()

    repo := NewBlocksRepository(s)
    ctx := context.Background()

    mock.ExpectBegin()
    mock.ExpectQuery("INSERT INTO block").WithArgs(uint64(1), sqlmock.AnyArg(), uint64(1)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(11))
    mock.ExpectExec("INSERT INTO block_attachment").WithArgs(uint64(11), uint64(5)).WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectCommit()

    row := sqlmock.NewRows([]string{"id", "note_id", "type", "position", "created_at", "updated_at", "text", "file_url"}).
        AddRow(11, 1, "attachment", 2.0, time.Now(), time.Now(), nil, "http://file")
    mock.ExpectQuery("SELECT b.id, b.note_id, b.type, b.position").WithArgs(uint64(11)).WillReturnRows(row)

    b, err := repo.CreateAttachmentBlock(ctx, 1, 2.0, 5, 1)
    assert.NoError(t, err)
    assert.NotNil(t, b)
    assert.Equal(t, uint64(11), b.ID)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBlocksRepository_GetBlocksByNoteIDForPositionCalc_Success(t *testing.T) {
    db, mock, s := setupTestDB(t)
    defer db.Close()

    repo := NewBlocksRepository(s)
    ctx := context.Background()

    rows := sqlmock.NewRows([]string{"id", "position"}).AddRow(2, 1.0).AddRow(3, 2.0)
    mock.ExpectQuery("SELECT id, position").WithArgs(uint64(1), uint64(0)).WillReturnRows(rows)

    info, err := repo.GetBlocksByNoteIDForPositionCalc(ctx, 1, 0)
    assert.NoError(t, err)
    if assert.Len(t, info, 2) {
        assert.Equal(t, uint64(2), info[0].ID)
    }
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBlocksRepository_UpdateBlockPosition_NotFound(t *testing.T) {
    db, mock, s := setupTestDB(t)
    defer db.Close()

    repo := NewBlocksRepository(s)
    ctx := context.Background()

    mock.ExpectQuery("UPDATE block").WithArgs(3.14, sqlmock.AnyArg(), uint64(999)).WillReturnError(sql.ErrNoRows)

    _, err := repo.UpdateBlockPosition(ctx, 999, 3.14)
    assert.Error(t, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBlocksRepository_DeleteBlock_NotFound(t *testing.T) {
    db, mock, s := setupTestDB(t)
    defer db.Close()

    repo := NewBlocksRepository(s)
    ctx := context.Background()

    mock.ExpectExec("DELETE FROM block WHERE id =").WithArgs(uint64(999)).WillReturnResult(sqlmock.NewResult(0, 0))

    err := repo.DeleteBlock(ctx, 999)
    assert.Error(t, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}
