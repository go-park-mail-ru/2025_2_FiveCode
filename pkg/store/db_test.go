package store

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestDBWrapper_QueryRowContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

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
	defer db.Close()

	wrapper := &dbWrapper{DB: db}
	ctx := context.Background()

	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))

	_, err = wrapper.ExecContext(ctx, "INSERT INTO table")
	assert.NoError(t, err)
}

func TestDBWrapper_QueryContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	wrapper := &dbWrapper{DB: db}
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	resRows, err := wrapper.QueryContext(ctx, "SELECT id")
	assert.NoError(t, err)
	defer resRows.Close()
}

func TestDBWrapper_BeginTx(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	wrapper := &dbWrapper{DB: db}
	ctx := context.Background()

	mock.ExpectBegin()

	tx, err := wrapper.BeginTx(ctx, nil)
	assert.NoError(t, err)
	assert.NotNil(t, tx)
}

func TestTxWrapper_QueryRowContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

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
	defer db.Close()

	mock.ExpectBegin()
	tx, _ := db.Begin()
	wrapper := &txWrapper{Tx: tx}
	ctx := context.Background()

	mock.ExpectExec("INSERT").WillReturnResult(sqlmock.NewResult(1, 1))

	_, err = wrapper.ExecContext(ctx, "INSERT INTO table")
	assert.NoError(t, err)
}

func TestTxWrapper_QueryContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, _ := db.Begin()
	wrapper := &txWrapper{Tx: tx}
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)

	resRows, err := wrapper.QueryContext(ctx, "SELECT id")
	assert.NoError(t, err)
	defer resRows.Close()
}

func TestTxWrapper_Commit(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

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
	defer db.Close()

	mock.ExpectBegin()
	tx, _ := db.Begin()
	wrapper := &txWrapper{Tx: tx}

	mock.ExpectRollback()

	err = wrapper.Rollback()
	assert.NoError(t, err)
}
