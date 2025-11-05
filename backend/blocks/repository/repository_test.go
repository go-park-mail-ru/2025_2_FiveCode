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

	s := &store.Store{
		Postgres: &store.PostgresDB{DB: db},
	}

	return db, mock, s
}

func TestBlocksRepository_CreateTextBlock(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewBlocksRepository(store)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO block`).
		WithArgs(uint64(1), 1.0, uint64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectExec(`INSERT INTO block_text`).
		WithArgs(uint64(1), "").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	blockRows := sqlmock.NewRows([]string{"id", "note_id", "type", "position", "created_at", "updated_at", "text", "file_url"}).
		AddRow(1, 1, "text", 1.0, time.Now(), time.Now(), "", nil)
	mock.ExpectQuery(`SELECT b.id, b.note_id, b.type, b.position`).
		WithArgs(uint64(1)).
		WillReturnRows(blockRows)

	formatRows := sqlmock.NewRows([]string{"id", "block_text_id", "start_offset", "end_offset", "bold", "italic", "underline", "strikethrough", "link", "font", "size"})
	mock.ExpectQuery(`SELECT btf.id, btf.block_text_id`).
		WithArgs(uint64(1)).
		WillReturnRows(formatRows)

	block, err := repo.CreateTextBlock(ctx, 1, 1.0, 1)
	assert.NoError(t, err)
	assert.NotNil(t, block)
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