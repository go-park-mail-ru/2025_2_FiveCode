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

func newTestRepo(db *sql.DB) *SharingRepository {
	return NewSharingRepository(&testStoreDB{DB: db})
}

func TestSharingRepository_AddCollaborator(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	
	permission := &models.NotePermission{
		NoteID:    1,
		GrantedBy: 1,
		GrantedTo: 2,
		Role:      models.RoleEditor,
	}

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"note_permission_id", "created_at", "updated_at"}).
			AddRow(1, time.Now(), time.Now())

		mock.ExpectQuery(`INSERT INTO note_permission`).
			WithArgs(permission.NoteID, permission.GrantedBy, permission.GrantedTo, permission.Role).
			WillReturnRows(rows)

		res, err := repo.AddCollaborator(ctx, permission)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, uint64(1), res.PermissionID)
	})
}

func TestSharingRepository_GetCollaboratorsByNoteID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	noteID := uint64(1)

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"note_permission_id", "note_id", "granted_by", "granted_to", "role", "created_at", "updated_at"}).
			AddRow(1, noteID, 1, 2, "editor", time.Now(), time.Now())

		mock.ExpectQuery(`SELECT (.+) FROM note_permission np`).
			WithArgs(noteID).
			WillReturnRows(rows)

		collaborators, err := repo.GetCollaboratorsByNoteID(ctx, noteID)
		assert.NoError(t, err)
		assert.Len(t, collaborators, 1)
		assert.Equal(t, models.RoleEditor, collaborators[0].Role)
	})
}

func TestSharingRepository_GetCollaboratorByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	permID := uint64(1)

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"note_permission_id", "note_id", "granted_by", "granted_to", "role", "created_at", "updated_at"}).
			AddRow(permID, 1, 1, 2, "editor", time.Now(), time.Now())

		mock.ExpectQuery(`SELECT (.+) FROM note_permission`).
			WithArgs(permID).
			WillReturnRows(rows)

		p, err := repo.GetCollaboratorByID(ctx, permID)
		assert.NoError(t, err)
		assert.Equal(t, permID, p.PermissionID)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(`SELECT (.+) FROM note_permission`).
			WithArgs(permID).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetCollaboratorByID(ctx, permID)
		assert.ErrorIs(t, err, constants.ErrNotFound)
	})
}

func TestSharingRepository_UpdateCollaboratorRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	permID := uint64(1)
	newRole := models.RoleViewer

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(`UPDATE note_permission`).
			WithArgs(newRole, permID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateCollaboratorRole(ctx, permID, newRole)
		assert.NoError(t, err)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectExec(`UPDATE note_permission`).
			WithArgs(newRole, permID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.UpdateCollaboratorRole(ctx, permID, newRole)
		assert.ErrorIs(t, err, constants.ErrNotFound)
	})
}

func TestSharingRepository_RemoveCollaborator(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	permID := uint64(1)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM note_permission`).
			WithArgs(permID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.RemoveCollaborator(ctx, permID)
		assert.NoError(t, err)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectExec(`DELETE FROM note_permission`).
			WithArgs(permID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.RemoveCollaborator(ctx, permID)
		assert.ErrorIs(t, err, constants.ErrNotFound)
	})
}

func TestSharingRepository_CheckCollaboratorExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	noteID := uint64(1)
	userID := uint64(2)

	t.Run("Exists", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
		mock.ExpectQuery(`SELECT EXISTS`).WithArgs(noteID, userID).WillReturnRows(rows)
		exists, err := repo.CheckCollaboratorExists(ctx, noteID, userID)
		assert.NoError(t, err)
		assert.True(t, exists)
	})
}

func TestSharingRepository_SetPublicAccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	noteID := uint64(1)
	role := models.RoleViewer

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(`UPDATE note SET public_access_level`).
			WithArgs(role, noteID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.SetPublicAccess(ctx, noteID, &role)
		assert.NoError(t, err)
	})

	t.Run("Disable", func(t *testing.T) {
		mock.ExpectExec(`UPDATE note SET public_access_level`).
			WithArgs(nil, noteID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.SetPublicAccess(ctx, noteID, nil)
		assert.NoError(t, err)
	})
}

func TestSharingRepository_GetPublicAccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	noteID := uint64(1)

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"public_access_level", "share_uuid"}).
			AddRow("viewer", "uuid")
		
		mock.ExpectQuery(`SELECT public_access_level, share_uuid`).
			WithArgs(noteID).
			WillReturnRows(rows)

		role, uuid, err := repo.GetPublicAccess(ctx, noteID)
		assert.NoError(t, err)
		assert.Equal(t, models.RoleViewer, *role)
		assert.Equal(t, "uuid", uuid)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(`SELECT public_access_level, share_uuid`).
			WithArgs(noteID).
			WillReturnError(sql.ErrNoRows)

		_, _, err := repo.GetPublicAccess(ctx, noteID)
		assert.ErrorIs(t, err, constants.ErrNotFound)
	})
}

func TestSharingRepository_GetNoteOwnerID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	noteID := uint64(1)

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"owner_id"}).AddRow(1)
		mock.ExpectQuery(`SELECT owner_id`).WithArgs(noteID).WillReturnRows(rows)
		id, err := repo.GetNoteOwnerID(ctx, noteID)
		assert.NoError(t, err)
		assert.Equal(t, uint64(1), id)
	})
}

func TestSharingRepository_CheckNoteAccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	noteID := uint64(1)
	userID := uint64(2)

	t.Run("Success_Editor", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"owner_id", "role"}).
			AddRow(1, "editor")

		mock.ExpectQuery(`SELECT (.+) FROM note n`).
			WithArgs(noteID, userID).
			WillReturnRows(rows)

		access, err := repo.CheckNoteAccess(ctx, noteID, userID)
		assert.NoError(t, err)
		assert.True(t, access.HasAccess)
		assert.True(t, access.CanEdit)
		assert.Equal(t, models.RoleEditor, access.Role)
	})

	t.Run("Success_Owner", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"owner_id", "role"}).
			AddRow(userID, nil)

		mock.ExpectQuery(`SELECT (.+) FROM note n`).
			WithArgs(noteID, userID).
			WillReturnRows(rows)

		access, err := repo.CheckNoteAccess(ctx, noteID, userID)
		assert.NoError(t, err)
		assert.True(t, access.IsOwner)
		assert.True(t, access.HasAccess)
		assert.True(t, access.CanEdit)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(`SELECT (.+) FROM note n`).
			WithArgs(noteID, userID).
			WillReturnError(sql.ErrNoRows)

		access, err := repo.CheckNoteAccess(ctx, noteID, userID)
		assert.NoError(t, err)
		assert.False(t, access.HasAccess)
	})
}

func TestSharingRepository_IsNoteOwner(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	noteID := uint64(1)
	userID := uint64(1)

	t.Run("Yes", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"owner_id"}).AddRow(userID)
		mock.ExpectQuery(`SELECT owner_id`).WithArgs(noteID).WillReturnRows(rows)
		isOwner, err := repo.IsNoteOwner(ctx, noteID, userID)
		assert.NoError(t, err)
		assert.True(t, isOwner)
	})
}

func TestSharingRepository_GetUserPermission(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	noteID := uint64(1)
	userID := uint64(2)

	t.Run("Found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"note_permission_id", "note_id", "granted_by", "granted_to", "role", "created_at", "updated_at"}).
			AddRow(1, noteID, 1, userID, "viewer", time.Now(), time.Now())
		
		mock.ExpectQuery(`SELECT (.+) FROM note_permission`).WithArgs(noteID, userID).WillReturnRows(rows)
		
		perm, err := repo.GetUserPermission(ctx, noteID, userID)
		assert.NoError(t, err)
		assert.NotNil(t, perm)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(`SELECT (.+) FROM note_permission`).WithArgs(noteID, userID).WillReturnError(sql.ErrNoRows)
		
		perm, err := repo.GetUserPermission(ctx, noteID, userID)
		assert.NoError(t, err)
		assert.Nil(t, perm)
	})
}

func TestSharingRepository_CanUserShare(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	noteID := uint64(1)
	userID := uint64(1)

	t.Run("OwnerCanShare", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"owner_id"}).AddRow(userID)
		mock.ExpectQuery(`SELECT owner_id`).WithArgs(noteID).WillReturnRows(rows)
		canShare, err := repo.CanUserShare(ctx, noteID, userID)
		assert.NoError(t, err)
		assert.True(t, canShare)
	})
}

func TestSharingRepository_UpdateIsSharedFlag(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a database connection", err)
	}
	defer db.Close()

	repo := newTestRepo(db)
	ctx := context.Background()
	noteID := uint64(1)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(`UPDATE note SET is_shared`).
			WithArgs(true, noteID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.UpdateIsSharedFlag(ctx, noteID, true)
		assert.NoError(t, err)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectExec(`UPDATE note SET is_shared`).
			WithArgs(true, noteID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.UpdateIsSharedFlag(ctx, noteID, true)
		assert.ErrorIs(t, err, constants.ErrNotFound)
	})
}
