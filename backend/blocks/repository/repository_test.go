package repository

import (
	"backend/models"
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

func TestBlocksRepository_CreateBlock(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewBlocksRepository(store)
	ctx := context.Background()

	mock.ExpectBegin()
	rows := sqlmock.NewRows([]string{"id", "note_id", "type", "position", "created_at", "updated_at", "last_edited_by"}).
		AddRow(1, 1, "text", 1.0, time.Now(), time.Now(), nil)
	mock.ExpectQuery(`INSERT INTO block`).
		WithArgs(1, "text", 1.0).
		WillReturnRows(rows)
	mock.ExpectExec(`INSERT INTO block_text`).
		WithArgs(1, "").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	block, err := repo.CreateBlock(ctx, 1, models.BlockTypeText, 1.0)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), block.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBlocksRepository_GetBlocksByNoteID(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewBlocksRepository(store)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "note_id", "type", "position", "created_at", "updated_at", "text", "formats"}).
		AddRow(1, 1, "text", 1.0, time.Now(), time.Now(), "Test", "[]")

	mock.ExpectQuery(`SELECT b.id, b.note_id, b.type, b.position`).
		WithArgs(1).
		WillReturnRows(rows)

	blocks, err := repo.GetBlocksByNoteID(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(blocks))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBlocksRepository_GetBlockByID(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewBlocksRepository(store)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "note_id", "type", "position", "created_at", "updated_at", "text"}).
		AddRow(1, 1, "text", 1.0, time.Now(), time.Now(), "Test")

	mock.ExpectQuery(`SELECT b.id, b.note_id, b.type, b.position`).
		WithArgs(1).
		WillReturnRows(rows)

	formatsRows := sqlmock.NewRows([]string{"id", "block_text_id", "start_offset", "end_offset", "bold", "italic", "underline", "strikethrough", "link", "font", "size"})
	mock.ExpectQuery(`SELECT btf.id, btf.block_text_id`).
		WithArgs(1).
		WillReturnRows(formatsRows)

	block, err := repo.GetBlockByID(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), block.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBlocksRepository_GetBlockByID_NotFound(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewBlocksRepository(store)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT b.id, b.note_id, b.type, b.position`).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	block, err := repo.GetBlockByID(ctx, 999)
	assert.Error(t, err)
	assert.Equal(t, namederrors.ErrNotFound, err)
	assert.Nil(t, block)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBlocksRepository_GetBlockByID_WithFormats(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewBlocksRepository(store)
	ctx := context.Background()

	blockRows := sqlmock.NewRows([]string{"id", "note_id", "type", "position", "created_at", "updated_at", "text"}).
		AddRow(1, 1, "text", 1.0, time.Now(), time.Now(), "Test text")

	mock.ExpectQuery(`SELECT b.id, b.note_id, b.type, b.position`).
		WithArgs(1).
		WillReturnRows(blockRows)

	formatRows := sqlmock.NewRows([]string{"id", "block_text_id", "start_offset", "end_offset", "bold", "italic", "underline", "strikethrough", "link", "font", "size"}).
		AddRow(1, 1, 0, 4, true, false, false, false, nil, "Inter", 12)

	mock.ExpectQuery(`SELECT btf.id, btf.block_text_id`).
		WithArgs(1).
		WillReturnRows(formatRows)

	block, err := repo.GetBlockByID(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), block.ID)
	assert.Equal(t, "Test text", block.Text)
	assert.Equal(t, 1, len(block.Formats))
	assert.True(t, block.Formats[0].Bold)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBlocksRepository_DeleteBlock(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewBlocksRepository(store)
	ctx := context.Background()

	mock.ExpectExec(`DELETE FROM block`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteBlock(ctx, 1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBlocksRepository_GetBlockNoteID(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewBlocksRepository(store)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"note_id"}).
		AddRow(1)

	mock.ExpectQuery(`SELECT note_id FROM block`).
		WithArgs(1).
		WillReturnRows(rows)

	noteID, err := repo.GetBlockNoteID(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), noteID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBlocksRepository_GetBlocksByNoteIDForPositionCalc(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewBlocksRepository(store)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "position"}).
		AddRow(1, 1.0).
		AddRow(2, 2.0)

	mock.ExpectQuery(`SELECT id, position`).
		WithArgs(1, 0).
		WillReturnRows(rows)

	blocks, err := repo.GetBlocksByNoteIDForPositionCalc(ctx, 1, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(blocks))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBlocksRepository_UpdateBlockText(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewBlocksRepository(store)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE block SET updated_at`).WithArgs(sqlmock.AnyArg(), uint64(1)).WillReturnResult(sqlmock.NewResult(0, 1))

	rows := sqlmock.NewRows([]string{"id"}).AddRow(uint64(1))
	mock.ExpectQuery(`UPDATE block_text`).WithArgs("text", sqlmock.AnyArg(), uint64(1)).WillReturnRows(rows)

	mock.ExpectExec(`DELETE FROM block_text_format`).WithArgs(uint64(1)).WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(`INSERT INTO block_text_format`).WithArgs(
		uint64(1), 0, 4, false, false, false, false, nil, "Inter", 12,
	).WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	blockRows := sqlmock.NewRows([]string{"id", "note_id", "type", "position", "created_at", "updated_at", "text"}).
		AddRow(1, 1, "text", 1.0, time.Now(), time.Now(), "text")
	mock.ExpectQuery(`SELECT b.id, b.note_id, b.type, b.position`).WithArgs(uint64(1)).WillReturnRows(blockRows)

	formatRows := sqlmock.NewRows([]string{"id", "block_text_id", "start_offset", "end_offset", "bold", "italic", "underline", "strikethrough", "link", "font", "size"})
	mock.ExpectQuery(`SELECT btf.id, btf.block_text_id`).WithArgs(uint64(1)).WillReturnRows(formatRows)

	formats := []models.BlockTextFormat{
		{StartOffset: 0, EndOffset: 4, Bold: false, Font: models.FontInter, Size: 12},
	}
	block, err := repo.UpdateBlockText(ctx, 1, "text", formats)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), block.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBlocksRepository_UpdateBlockPosition(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewBlocksRepository(store)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "note_id", "type", "position", "created_at", "updated_at"}).
		AddRow(1, 1, "text", 2.5, time.Now(), time.Now())

	mock.ExpectQuery(`UPDATE block`).
		WithArgs(2.5, sqlmock.AnyArg(), uint64(1)).
		WillReturnRows(rows)

	block, err := repo.UpdateBlockPosition(ctx, 1, 2.5)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), block.ID)
	assert.Equal(t, 2.5, block.Position)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBlocksRepository_UpdateBlockPosition_NotFound(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewBlocksRepository(store)
	ctx := context.Background()

	mock.ExpectQuery(`UPDATE block`).
		WithArgs(2.5, sqlmock.AnyArg(), uint64(999)).
		WillReturnError(sql.ErrNoRows)

	block, err := repo.UpdateBlockPosition(ctx, 999, 2.5)
	assert.Error(t, err)
	assert.Equal(t, namederrors.ErrNotFound, err)
	assert.Nil(t, block)
	assert.NoError(t, mock.ExpectationsWereMet())
}
