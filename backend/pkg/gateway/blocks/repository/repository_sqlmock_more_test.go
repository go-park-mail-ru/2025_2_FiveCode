package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"backend/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestBlocksRepository_UpdateBlockText_Success(t *testing.T) {
    db, mock, s := setupTestDB(t)
    defer db.Close()

    repo := NewBlocksRepository(s)
    ctx := context.Background()

    mock.ExpectBegin()
    // update block timestamp exec
    mock.ExpectExec("UPDATE block SET updated_at").WithArgs(sqlmock.AnyArg(), uint64(1)).WillReturnResult(sqlmock.NewResult(1, 1))
    // update block_text returning id
    mock.ExpectQuery("UPDATE block_text").WithArgs("new text", sqlmock.AnyArg(), uint64(1)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))
    // delete old formats
    mock.ExpectExec("DELETE FROM block_text_format WHERE block_text_id =").WithArgs(uint64(5)).WillReturnResult(sqlmock.NewResult(1, 1))
    // insert new format
    mock.ExpectExec("INSERT INTO block_text_format").WithArgs(uint64(5), 0, 3, true, false, false, false, nil, models.FontInter, 12).WillReturnResult(sqlmock.NewResult(1, 1))
    mock.ExpectCommit()

    // After commit, GetBlockByID will be called: return a row
    blockRow := sqlmock.NewRows([]string{"id", "note_id", "type", "position", "created_at", "updated_at", "text", "file_url"}).
        AddRow(1, 1, "text", 1.0, time.Now(), time.Now(), "new text", nil)
    mock.ExpectQuery("SELECT b.id, b.note_id, b.type, b.position").WithArgs(uint64(1)).WillReturnRows(blockRow)

    // formats query for GetBlockByID
    formatRow := sqlmock.NewRows([]string{"id", "block_text_id", "start_offset", "end_offset", "bold", "italic", "underline", "strikethrough", "link", "font", "size"}).
        AddRow(1, 5, 0, 3, true, false, false, false, nil, models.FontInter, 12)
    mock.ExpectQuery("SELECT btf.id, btf.block_text_id").WithArgs(uint64(1)).WillReturnRows(formatRow)

    formats := []models.BlockTextFormat{{StartOffset:0, EndOffset:3, Bold:true, Font:models.FontInter, Size:12}}
    _, err := repo.UpdateBlockText(ctx, 1, "new text", formats)
    assert.NoError(t, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBlocksRepository_GetBlocksByNoteID_WithFormats(t *testing.T) {
    db, mock, s := setupTestDB(t)
    defer db.Close()

    repo := NewBlocksRepository(s)
    ctx := context.Background()

    formats := []map[string]interface{}{{"id":1, "start_offset":0, "end_offset":3, "bold":true, "italic":false, "underline":false, "strikethrough":false, "link":nil, "font":string(models.FontInter), "size":12}}
    bts, _ := json.Marshal(formats)

    rows := sqlmock.NewRows([]string{"id", "note_id", "type", "position", "created_at", "updated_at", "text", "file_url", "formats"}).
        AddRow(2, 1, "text", 1.0, time.Now(), time.Now(), "hello", nil, bts)

    mock.ExpectQuery("SELECT b.id, b.note_id, b.type, b.position").WithArgs(uint64(1)).WillReturnRows(rows)

    res, err := repo.GetBlocksByNoteID(ctx, 1)
    assert.NoError(t, err)
    if assert.Len(t, res, 1) {
        assert.Equal(t, uint64(2), res[0].ID)
        assert.Len(t, res[0].Formats, 1)
        assert.True(t, res[0].Formats[0].Bold)
    }
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBlocksRepository_UpdateBlockPosition_Success(t *testing.T) {
    db, mock, s := setupTestDB(t)
    defer db.Close()

    repo := NewBlocksRepository(s)
    ctx := context.Background()

    // RETURNING row for update
    updatedRow := sqlmock.NewRows([]string{"id", "note_id", "type", "position", "created_at", "updated_at"}).AddRow(3, 1, "text", 3.14, time.Now(), time.Now())
    mock.ExpectQuery("UPDATE block").WithArgs(3.14, sqlmock.AnyArg(), uint64(3)).WillReturnRows(updatedRow)

    blk, err := repo.UpdateBlockPosition(ctx, 3, 3.14)
    assert.NoError(t, err)
    assert.Equal(t, uint64(3), blk.ID)
    assert.Equal(t, 3.14, blk.Position)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBlocksRepository_UpdateBlockText_ErrorBranch(t *testing.T) {
    db, mock, s := setupTestDB(t)
    defer db.Close()

    repo := NewBlocksRepository(s)
    ctx := context.Background()

    mock.ExpectBegin()
    // fail to update block timestamp
    mock.ExpectExec("UPDATE block SET updated_at").WithArgs(sqlmock.AnyArg(), uint64(1)).WillReturnError(sql.ErrConnDone)

    _, err := repo.UpdateBlockText(ctx, 1, "x", nil)
    assert.Error(t, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBlocksRepository_DeleteBlock_Success(t *testing.T) {
    db, mock, s := setupTestDB(t)
    defer db.Close()

    repo := NewBlocksRepository(s)
    ctx := context.Background()

    mock.ExpectExec("DELETE FROM block WHERE id =").WithArgs(uint64(2)).WillReturnResult(sqlmock.NewResult(1, 1))

    err := repo.DeleteBlock(ctx, 2)
    assert.NoError(t, err)
    assert.NoError(t, mock.ExpectationsWereMet())
}
