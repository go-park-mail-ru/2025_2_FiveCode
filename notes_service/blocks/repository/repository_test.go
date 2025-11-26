package repository

import (
	"backend/notes_service/internal/constants"
	"backend/notes_service/internal/models"
	"backend/pkg/store"
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
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

func newTestRepo(db *sql.DB) *BlocksRepository {
	return NewBlocksRepository(&testStoreDB{DB: db})
}

func expectGetBlockByID(mock sqlmock.Sqlmock, blockID uint64, noteID uint64, blockType string) {
	baseRows := sqlmock.NewRows([]string{"id", "note_id", "type", "position", "created_at", "updated_at"}).
		AddRow(blockID, noteID, blockType, 1.0, time.Now(), time.Now())
	mock.ExpectQuery(`SELECT (.+) FROM block`).
		WithArgs(pq.Array([]uint64{blockID})).
		WillReturnRows(baseRows)

	if blockType == models.BlockTypeText {
		textRows := sqlmock.NewRows([]string{"block_id", "text", "formats"}).
			AddRow(blockID, "some text", []byte("[]"))
		mock.ExpectQuery(`SELECT (.+) FROM block_text`).
			WithArgs(pq.Array([]uint64{blockID})).
			WillReturnRows(textRows)
	}

	if blockType == models.BlockTypeCode {
		codeRows := sqlmock.NewRows([]string{"block_id", "code_text", "language"}).
			AddRow(blockID, "console.log()", "javascript")
		mock.ExpectQuery(`SELECT (.+) FROM block_code`).
			WithArgs(pq.Array([]uint64{blockID})).
			WillReturnRows(codeRows)
	}

	if blockType == models.BlockTypeAttachment {
		attRows := sqlmock.NewRows([]string{"block_id", "file_id", "caption"}).
			AddRow(blockID, 123, "caption")
		mock.ExpectQuery(`SELECT (.+) FROM block_attachment`).
			WithArgs(pq.Array([]uint64{blockID})).
			WillReturnRows(attRows)
	}
}

func TestBlocksRepository_CreateTextBlock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)
	blockID := uint64(100)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"id"}).AddRow(blockID)
		mock.ExpectQuery(`INSERT INTO block`).
			WithArgs(noteID, 1.0, userID, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(rows)

		mock.ExpectExec(`INSERT INTO block_text`).
			WithArgs(blockID, "", sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		expectGetBlockByID(mock, blockID, noteID, models.BlockTypeText)

		block, err := repo.CreateTextBlock(ctx, noteID, 1.0, userID)
		assert.NoError(t, err)
		assert.NotNil(t, block)
		assert.Equal(t, blockID, block.BaseBlock.ID)
	})
}

func TestBlocksRepository_CreateAttachmentBlock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)
	blockID := uint64(100)
	fileID := uint64(123)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"id"}).AddRow(blockID)
		mock.ExpectQuery(`INSERT INTO block`).
			WithArgs(noteID, 1.0, userID, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(rows)

		mock.ExpectExec(`INSERT INTO block_attachment`).
			WithArgs(blockID, fileID, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		expectGetBlockByID(mock, blockID, noteID, models.BlockTypeAttachment)

		block, err := repo.CreateAttachmentBlock(ctx, noteID, 1.0, fileID, userID)
		assert.NoError(t, err)
		assert.NotNil(t, block)
		assert.Equal(t, blockID, block.BaseBlock.ID)
	})
}

func TestBlocksRepository_CreateCodeBlock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	userID := uint64(1)
	noteID := uint64(10)
	blockID := uint64(100)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()

		rows := sqlmock.NewRows([]string{"id"}).AddRow(blockID)
		mock.ExpectQuery(`INSERT INTO block`).
			WithArgs(noteID, 1.0, userID, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(rows)

		mock.ExpectExec(`INSERT INTO block_code`).
			WithArgs(blockID, "javascript", "", sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		expectGetBlockByID(mock, blockID, noteID, models.BlockTypeCode)

		block, err := repo.CreateCodeBlock(ctx, noteID, 1.0, userID)
		assert.NoError(t, err)
		assert.NotNil(t, block)
		assert.Equal(t, blockID, block.BaseBlock.ID)
	})
}

func TestBlocksRepository_UpdateCodeBlock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	blockID := uint64(100)
	noteID := uint64(10)
	language := "go"
	codeText := "fmt.Println()"

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectExec(`UPDATE block SET updated_at`).
			WithArgs(sqlmock.AnyArg(), blockID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(`INSERT INTO block_code`).
			WithArgs(blockID, language, codeText, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		expectGetBlockByID(mock, blockID, noteID, models.BlockTypeCode)

		block, err := repo.UpdateCodeBlock(ctx, blockID, language, codeText)
		assert.NoError(t, err)
		assert.NotNil(t, block)
	})
}

func TestBlocksRepository_UpdateBlockText(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	blockID := uint64(100)
	noteID := uint64(10)
	text := "new text"
	formats := []models.BlockTextFormat{
		{StartOffset: 0, EndOffset: 5, Bold: true},
	}

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectExec(`UPDATE block SET updated_at`).
			WithArgs(sqlmock.AnyArg(), blockID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		blockTextID := uint64(500)
		mock.ExpectQuery(`UPDATE block_text`).
			WithArgs(text, sqlmock.AnyArg(), blockID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(blockTextID))

		mock.ExpectExec(`DELETE FROM block_text_format`).
			WithArgs(blockTextID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(`INSERT INTO block_text_format`).
			WithArgs(blockTextID, formats[0].StartOffset, formats[0].EndOffset, formats[0].Bold, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		expectGetBlockByID(mock, blockID, noteID, models.BlockTypeText)

		block, err := repo.UpdateBlockText(ctx, blockID, text, formats)
		assert.NoError(t, err)
		assert.NotNil(t, block)
	})
}

func TestBlocksRepository_UpdateBlockPosition(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	blockID := uint64(100)
	noteID := uint64(10)
	position := 2.5

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(`UPDATE block SET position`).
			WithArgs(position, sqlmock.AnyArg(), blockID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		expectGetBlockByID(mock, blockID, noteID, models.BlockTypeText)

		block, err := repo.UpdateBlockPosition(ctx, blockID, position)
		assert.NoError(t, err)
		assert.NotNil(t, block)
	})
}

func TestBlocksRepository_DeleteBlock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	blockID := uint64(100)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM block`).
			WithArgs(blockID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteBlock(ctx, blockID)
		assert.NoError(t, err)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM block`).
			WithArgs(blockID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.DeleteBlock(ctx, blockID)
		assert.Error(t, err)
		assert.Equal(t, constants.ErrNotFound, err)
	})
}

func TestBlocksRepository_GetBlockNoteID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	blockID := uint64(100)
	noteID := uint64(10)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(`SELECT note_id FROM block`).
			WithArgs(blockID).
			WillReturnRows(sqlmock.NewRows([]string{"note_id"}).AddRow(noteID))

		id, err := repo.GetBlockNoteID(ctx, blockID)
		assert.NoError(t, err)
		assert.Equal(t, noteID, id)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(`SELECT note_id FROM block`).
			WithArgs(blockID).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetBlockNoteID(ctx, blockID)
		assert.ErrorIs(t, err, constants.ErrNotFound)
	})
}

func TestBlocksRepository_GetBlocksByNoteID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	noteID := uint64(10)
	blockID := uint64(100)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id FROM block`).
			WithArgs(noteID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(blockID))

		baseRows := sqlmock.NewRows([]string{"id", "note_id", "type", "position", "created_at", "updated_at"}).
			AddRow(blockID, noteID, models.BlockTypeText, 1.0, time.Now(), time.Now())
		mock.ExpectQuery(`SELECT (.+) FROM block`).
			WithArgs(pq.Array([]uint64{blockID})).
			WillReturnRows(baseRows)

		textRows := sqlmock.NewRows([]string{"block_id", "text", "formats"}).
			AddRow(blockID, "text", []byte("[]"))
		mock.ExpectQuery(`SELECT (.+) FROM block_text`).
			WithArgs(pq.Array([]uint64{blockID})).
			WillReturnRows(textRows)

		mock.ExpectQuery(`SELECT (.+) FROM block_code`).
			WithArgs(pq.Array([]uint64{})).
			WillReturnRows(sqlmock.NewRows([]string{}))

		mock.ExpectQuery(`SELECT (.+) FROM block_attachment`).
			WithArgs(pq.Array([]uint64{})).
			WillReturnRows(sqlmock.NewRows([]string{}))

		blocks, err := repo.GetBlocksByNoteID(ctx, noteID)
		assert.NoError(t, err)
		assert.Len(t, blocks, 1)
	})
}

func TestBlocksRepository_GetBlocksByNoteIDForPositionCalc(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	noteID := uint64(10)
	excludeBlockID := uint64(100)

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "position"}).
			AddRow(101, 2.0).
			AddRow(102, 3.0)

		mock.ExpectQuery(`SELECT id, position FROM block`).
			WithArgs(noteID, excludeBlockID).
			WillReturnRows(rows)

		blocks, err := repo.GetBlocksByNoteIDForPositionCalc(ctx, noteID, excludeBlockID)
		assert.NoError(t, err)
		assert.Len(t, blocks, 2)
	})
}
