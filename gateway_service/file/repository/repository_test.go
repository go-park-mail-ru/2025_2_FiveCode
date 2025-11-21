package repository

import (
	namederrors "backend/named_errors"
	store2 "backend/pkg/store"
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestFileRepository_extractObjectNameFromURL_and_splitURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
		wantErr  bool
	}{
		{name: "simple URL", url: "http://example.com/bucket/file.txt", expected: "file.txt", wantErr: false},
		{name: "URL with path", url: "https://minio.example.com:9000/bucket/path/to/file.txt", expected: "file.txt", wantErr: false},
		{name: "URL with query", url: "http://example.com/bucket/file.txt?param=value", expected: "file.txt?param=value", wantErr: false},
		{name: "invalid URL", url: "invalid", expected: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractObjectNameFromURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}

	// splitURL specific checks
	parts := splitURL("http://example.com/bucket/file.txt")
	assert.GreaterOrEqual(t, len(parts), 2)
	assert.Equal(t, "file.txt", parts[len(parts)-1])
}

func setupTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *store2.Store) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}

	store := &store2.Store{
		Postgres: &store2.PostgresDB{DB: db},
	}

	return db, mock, store
}

func TestFileRepository_SaveFile(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewFileRepository(store)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "url", "mime_type", "size_bytes", "width", "height", "created_at", "updated_at"}).
		AddRow(1, "http://example.com/file.txt", "text/plain", 100, nil, nil, time.Now(), time.Now())

	mock.ExpectQuery(`INSERT INTO file`).
		WithArgs("http://example.com/file.txt", "text/plain", int64(100), nil, nil).
		WillReturnRows(rows)

	file, err := repo.SaveFile(ctx, "http://example.com/file.txt", "text/plain", 100, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), file.ID)
	assert.Equal(t, "http://example.com/file.txt", file.URL)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFileRepository_GetFileByID(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewFileRepository(store)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "url", "mime_type", "size_bytes", "width", "height", "created_at", "updated_at"}).
		AddRow(1, "http://example.com/file.txt", "text/plain", 100, nil, nil, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT id, url, mime_type, size_bytes, width, height, created_at, updated_at`).
		WithArgs(1).
		WillReturnRows(rows)

	file, err := repo.GetFileByID(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), file.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// File repository tests
func TestFileRepository_SaveFile_Success(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewFileRepository(store)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "url", "mime_type", "size_bytes", "width", "height", "created_at", "updated_at"}).
		AddRow(1, "http://ex/a", "text/plain", 5, nil, nil, time.Now(), time.Now())

	mock.ExpectQuery("INSERT INTO file").WithArgs("http://ex/a", "text/plain", int64(5), nil, nil).WillReturnRows(rows)

	f, err := repo.SaveFile(ctx, "http://ex/a", "text/plain", 5, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), f.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFileRepository_GetFileByID_NotFound(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewFileRepository(store)
	ctx := context.Background()

	mock.ExpectQuery("SELECT id, url, mime_type").WithArgs(uint64(999)).WillReturnError(sql.ErrNoRows)

	_, err := repo.GetFileByID(ctx, 999)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFileRepository_DeleteFile_SuccessAndNotFound(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewFileRepository(store)
	ctx := context.Background()

	// success
	mock.ExpectExec("DELETE FROM file WHERE id =").WithArgs(uint64(2)).WillReturnResult(sqlmock.NewResult(1, 1))
	err := repo.DeleteFile(ctx, 2)
	assert.NoError(t, err)

	// not found
	mock.ExpectExec("DELETE FROM file WHERE id =").WithArgs(uint64(999)).WillReturnResult(sqlmock.NewResult(0, 0))
	err = repo.DeleteFile(ctx, 999)
	assert.Error(t, err)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFileRepository_DeleteFile(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewFileRepository(store)
	ctx := context.Background()

	mock.ExpectExec(`DELETE FROM file`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteFile(ctx, 1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFileRepository_DeleteFile_NotFound(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewFileRepository(store)
	ctx := context.Background()

	mock.ExpectExec(`DELETE FROM file`).
		WithArgs(999).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.DeleteFile(ctx, 999)
	assert.Error(t, err)
	assert.Equal(t, namederrors.ErrNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFileRepository_SaveFile_WithDimensions(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewFileRepository(store)
	ctx := context.Background()

	width := 1920
	height := 1080
	rows := sqlmock.NewRows([]string{"id", "url", "mime_type", "size_bytes", "width", "height", "created_at", "updated_at"}).
		AddRow(1, "http://example.com/image.jpg", "image/jpeg", 1000, 1920, 1080, time.Now(), time.Now())

	mock.ExpectQuery(`INSERT INTO file`).
		WithArgs("http://example.com/image.jpg", "image/jpeg", int64(1000), 1920, 1080).
		WillReturnRows(rows)

	file, err := repo.SaveFile(ctx, "http://example.com/image.jpg", "image/jpeg", 1000, &width, &height)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), file.ID)
	assert.NotNil(t, file.Width)
	assert.NotNil(t, file.Height)
	assert.Equal(t, 1920, *file.Width)
	assert.Equal(t, 1080, *file.Height)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFileRepository_GetFileByID_WithDimensions(t *testing.T) {
	db, mock, store := setupTestDB(t)
	defer db.Close()

	repo := NewFileRepository(store)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "url", "mime_type", "size_bytes", "width", "height", "created_at", "updated_at"}).
		AddRow(1, "http://example.com/image.jpg", "image/jpeg", 1000, 1920, 1080, time.Now(), time.Now())

	mock.ExpectQuery(`SELECT id, url, mime_type, size_bytes, width, height, created_at, updated_at`).
		WithArgs(1).
		WillReturnRows(rows)

	file, err := repo.GetFileByID(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), file.ID)
	assert.NotNil(t, file.Width)
	assert.NotNil(t, file.Height)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFileRepository_extractObjectNameFromURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
		wantErr  bool
	}{
		{
			name:     "simple URL",
			url:      "http://example.com/bucket/file.txt",
			expected: "file.txt",
			wantErr:  false,
		},
		{
			name:     "URL with path",
			url:      "https://minio.example.com:9000/bucket/path/to/file.txt",
			expected: "file.txt",
			wantErr:  false,
		},
		{
			name:     "URL with query",
			url:      "http://example.com/bucket/file.txt?param=value",
			expected: "file.txt?param=value",
			wantErr:  false,
		},
		{
			name:     "invalid URL",
			url:      "invalid",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractObjectNameFromURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFileRepository_splitURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		minParts int
		lastPart string
	}{
		{
			name:     "simple URL",
			url:      "http://example.com/bucket/file.txt",
			minParts: 2,
			lastPart: "file.txt",
		},
		{
			name:     "URL with port",
			url:      "https://minio.example.com:9000/bucket/file.txt",
			minParts: 2,
			lastPart: "file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := splitURL(tt.url)
			assert.GreaterOrEqual(t, len(parts), tt.minParts)
			if len(parts) > 0 {
				assert.Equal(t, tt.lastPart, parts[len(parts)-1])
			}
		})
	}
}
