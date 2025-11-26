package repository

import (
	"backend/gateway_service/internal/constants"
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestFileRepository_SaveFile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := NewFileRepository(db, nil)
	ctx := context.Background()

	url := "http://example.com/test.jpg"
	mimeType := "image/jpeg"
	sizeBytes := int64(1024)
	width := 100
	height := 100

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "url", "mime_type", "size_bytes", "width", "height", "created_at", "updated_at"}).
			AddRow(1, url, mimeType, sizeBytes, width, height, time.Now(), time.Now())

		mock.ExpectQuery(`INSERT INTO file`).
			WithArgs(url, mimeType, sizeBytes, width, height, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(rows)

		file, err := repo.SaveFile(ctx, url, mimeType, sizeBytes, &width, &height)
		assert.NoError(t, err)
		assert.NotNil(t, file)
		assert.Equal(t, url, file.URL)
		assert.Equal(t, &width, file.Width)
	})

	t.Run("Success_NullDimensions", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "url", "mime_type", "size_bytes", "width", "height", "created_at", "updated_at"}).
			AddRow(1, url, mimeType, sizeBytes, nil, nil, time.Now(), time.Now())

		mock.ExpectQuery(`INSERT INTO file`).
			WithArgs(url, mimeType, sizeBytes, nil, nil, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(rows)

		file, err := repo.SaveFile(ctx, url, mimeType, sizeBytes, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, file)
		assert.Nil(t, file.Width)
		assert.Nil(t, file.Height)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(`INSERT INTO file`).
			WillReturnError(errors.New("db error"))

		_, err := repo.SaveFile(ctx, url, mimeType, sizeBytes, &width, &height)
		assert.Error(t, err)
	})
}

func TestFileRepository_GetFileByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := NewFileRepository(db, nil)
	ctx := context.Background()
	fileID := uint64(1)

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "url", "mime_type", "size_bytes", "width", "height", "created_at", "updated_at"}).
			AddRow(fileID, "http://example.com/test.jpg", "image/jpeg", 1024, 100, 100, time.Now(), time.Now())

		mock.ExpectQuery(`SELECT (.+) FROM file`).
			WithArgs(fileID).
			WillReturnRows(rows)

		file, err := repo.GetFileByID(ctx, fileID)
		assert.NoError(t, err)
		assert.NotNil(t, file)
		assert.Equal(t, fileID, file.ID)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(`SELECT (.+) FROM file`).
			WithArgs(fileID).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetFileByID(ctx, fileID)
		assert.Error(t, err)
		assert.Equal(t, constants.ErrNotFound, err)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(`SELECT (.+) FROM file`).
			WithArgs(fileID).
			WillReturnError(errors.New("db error"))

		_, err := repo.GetFileByID(ctx, fileID)
		assert.Error(t, err)
	})
}

func TestFileRepository_DeleteFile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := NewFileRepository(db, nil)
	ctx := context.Background()
	fileID := uint64(1)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM file`).
			WithArgs(fileID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.DeleteFile(ctx, fileID)
		assert.NoError(t, err)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM file`).
			WithArgs(fileID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.DeleteFile(ctx, fileID)
		assert.Error(t, err)
		assert.Equal(t, constants.ErrNotFound, err)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM file`).
			WithArgs(fileID).
			WillReturnError(errors.New("db error"))

		err := repo.DeleteFile(ctx, fileID)
		assert.Error(t, err)
	})
}
