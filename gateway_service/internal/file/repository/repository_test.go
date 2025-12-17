package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"backend/gateway_service/internal/constants"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestFileRepository_SaveFile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

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
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

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

	t.Run("WithNullDimensions", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "url", "mime_type", "size_bytes", "width", "height", "created_at", "updated_at"}).
			AddRow(fileID, "http://example.com/test.jpg", "image/jpeg", 1024, nil, nil, time.Now(), time.Now())

		mock.ExpectQuery(`SELECT (.+) FROM file`).
			WithArgs(fileID).
			WillReturnRows(rows)

		file, err := repo.GetFileByID(ctx, fileID)
		assert.NoError(t, err)
		assert.NotNil(t, file)
		assert.Nil(t, file.Width)
		assert.Nil(t, file.Height)
	})
}

func TestFileRepository_DeleteFile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

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

	t.Run("RowsAffectedError", func(t *testing.T) {
		result := &errorResult{err: errors.New("rows affected error")}
		mock.ExpectExec(`DELETE FROM file`).
			WithArgs(fileID).
			WillReturnResult(result)

		err := repo.DeleteFile(ctx, fileID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get rows affected")
	})
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

func TestFileRepository_GetIcons(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

	repo := NewFileRepository(db, nil)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "url"}).
			AddRow(1, "http://example.com/icons/icon1.svg").
			AddRow(2, "http://example.com/icons/icon2.svg")

		mock.ExpectQuery(`SELECT id, url`).
			WillReturnRows(rows)

		icons, err := repo.GetIcons(ctx)
		assert.NoError(t, err)
		assert.Len(t, icons, 2)
		assert.Equal(t, uint64(1), icons[0].ID)
		assert.Equal(t, "icon1.svg", icons[0].Name)
		assert.Equal(t, "http://example.com/icons/icon1.svg", icons[0].URL)
	})

	t.Run("EmptyResults", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "url"})

		mock.ExpectQuery(`SELECT id, url`).
			WillReturnRows(rows)

		icons, err := repo.GetIcons(ctx)
		assert.NoError(t, err)
		assert.Len(t, icons, 0)
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, url`).
			WillReturnError(errors.New("database error"))

		_, err := repo.GetIcons(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to query icons")
	})

	t.Run("ScanError", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "url"}).
			AddRow("invalid", "http://example.com/icons/icon1.svg")

		mock.ExpectQuery(`SELECT id, url`).
			WillReturnRows(rows)

		_, err := repo.GetIcons(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to scan icon")
	})

	t.Run("RowsError", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "url"}).
			AddRow(1, "http://example.com/icons/icon1.svg").
			RowError(0, errors.New("row error"))

		mock.ExpectQuery(`SELECT id, url`).
			WillReturnRows(rows)

		_, err := repo.GetIcons(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to iterate icons")
	})
}

func TestFileRepository_GetHeaders(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

	repo := NewFileRepository(db, nil)
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "url"}).
			AddRow(1, "http://example.com/headers/header1.jpg").
			AddRow(2, "http://example.com/headers/header2.jpg")

		mock.ExpectQuery(`SELECT id, url`).
			WillReturnRows(rows)

		headers, err := repo.GetHeaders(ctx)
		assert.NoError(t, err)
		assert.Len(t, headers, 2)
		assert.Equal(t, uint64(1), headers[0].ID)
		assert.Equal(t, "header1.jpg", headers[0].Name)
		assert.Equal(t, "http://example.com/headers/header1.jpg", headers[0].URL)
	})

	t.Run("EmptyResults", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "url"})

		mock.ExpectQuery(`SELECT id, url`).
			WillReturnRows(rows)

		headers, err := repo.GetHeaders(ctx)
		assert.NoError(t, err)
		assert.Len(t, headers, 0)
	})

	t.Run("DatabaseError", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, url`).
			WillReturnError(errors.New("database error"))

		_, err := repo.GetHeaders(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to query headers")
	})

	t.Run("ScanError", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "url"}).
			AddRow("invalid", "http://example.com/headers/header1.jpg")

		mock.ExpectQuery(`SELECT id, url`).
			WillReturnRows(rows)

		_, err := repo.GetHeaders(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to scan header")
	})

	t.Run("RowsError", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "url"}).
			AddRow(1, "http://example.com/headers/header1.jpg").
			RowError(0, errors.New("row error"))

		mock.ExpectQuery(`SELECT id, url`).
			WillReturnRows(rows)

		_, err := repo.GetHeaders(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to iterate headers")
	})
}
