package store

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestDBWrapper_QueryRowContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

	wrapper := &dbWrapper{DB: db}
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	var id int
	err = wrapper.QueryRowContext(ctx, "SELECT id").Scan(&id)
	assert.NoError(t, err)
	assert.Equal(t, 1, id)
}

func TestDBWrapper_ExecContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

	wrapper := &dbWrapper{DB: db}
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))

		_, err = wrapper.ExecContext(ctx, "INSERT INTO table")
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		mock.ExpectExec("INSERT").WillReturnError(errors.New("exec error"))

		_, err = wrapper.ExecContext(ctx, "INSERT INTO table")
		assert.Error(t, err)
	})
}

func TestDBWrapper_QueryContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

	wrapper := &dbWrapper{DB: db}
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
		mock.ExpectQuery("SELECT").WillReturnRows(rows)

		resRows, err := wrapper.QueryContext(ctx, "SELECT id")
		assert.NoError(t, err)
		defer func() {
			if err := resRows.Close(); err != nil {
				t.Logf("failed to close rows: %v", err)
			}
		}()
	})

	t.Run("Error", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WillReturnError(errors.New("query error"))

		_, err := wrapper.QueryContext(ctx, "SELECT id")
		assert.Error(t, err)
	})
}

func TestDBWrapper_BeginTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

	wrapper := &dbWrapper{DB: db}
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()

		tx, err := wrapper.BeginTx(ctx, nil)
		assert.NoError(t, err)
		assert.NotNil(t, tx)
	})

	t.Run("Error", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(errors.New("begin error"))

		_, err := wrapper.BeginTx(ctx, nil)
		assert.Error(t, err)
	})
}

func TestTxWrapper_QueryRowContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

	mock.ExpectBegin()
	tx, _ := db.Begin()
	wrapper := &txWrapper{Tx: tx}
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	var id int
	err = wrapper.QueryRowContext(ctx, "SELECT id").Scan(&id)
	assert.NoError(t, err)
	assert.Equal(t, 1, id)
}

func TestTxWrapper_ExecContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

	mock.ExpectBegin()
	tx, _ := db.Begin()
	wrapper := &txWrapper{Tx: tx}
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))

		_, err = wrapper.ExecContext(ctx, "INSERT INTO table")
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		mock.ExpectExec("INSERT").WillReturnError(errors.New("exec error"))

		_, err = wrapper.ExecContext(ctx, "INSERT INTO table")
		assert.Error(t, err)
	})
}

func TestTxWrapper_QueryContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

	mock.ExpectBegin()
	tx, _ := db.Begin()
	wrapper := &txWrapper{Tx: tx}
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
		mock.ExpectQuery("SELECT").WillReturnRows(rows)

		resRows, err := wrapper.QueryContext(ctx, "SELECT id")
		assert.NoError(t, err)
		defer func() {
			if err := resRows.Close(); err != nil {
				t.Logf("failed to close rows: %v", err)
			}
		}()
	})

	t.Run("Error", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WillReturnError(errors.New("query error"))

		_, err := wrapper.QueryContext(ctx, "SELECT id")
		assert.Error(t, err)
	})
}

func TestTxWrapper_Commit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

	mock.ExpectBegin()
	tx, _ := db.Begin()
	wrapper := &txWrapper{Tx: tx}

	mock.ExpectCommit()

	err = wrapper.Commit()
	assert.NoError(t, err)
}

func TestTxWrapper_Rollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close rows: %v", err)
		}
	}()

	mock.ExpectBegin()
	tx, _ := db.Begin()
	wrapper := &txWrapper{Tx: tx}

	mock.ExpectRollback()

	err = wrapper.Rollback()
	assert.NoError(t, err)
}
