package usecase

import (
	"backend/models"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type nu_fakeRepo struct{ owner uint64 }
func (r *nu_fakeRepo) GetNotes(ctx context.Context, userID uint64) ([]models.Note, error) { return []models.Note{{ID:1, OwnerID:r.owner}}, nil }
func (r *nu_fakeRepo) CreateNote(ctx context.Context, userID uint64) (*models.Note, error) { return &models.Note{ID:2, OwnerID:r.owner}, nil }
func (r *nu_fakeRepo) GetNoteById(ctx context.Context, noteID uint64, userID uint64) (*models.Note, error) { return &models.Note{ID:noteID, OwnerID:r.owner}, nil }
func (r *nu_fakeRepo) UpdateNote(ctx context.Context, noteID uint64, title *string, isArchived *bool) (*models.Note, error) { return &models.Note{ID:noteID, OwnerID:r.owner, Title: *title}, nil }
func (r *nu_fakeRepo) DeleteNote(ctx context.Context, noteID uint64) error { return nil }
func (r *nu_fakeRepo) AddFavorite(ctx context.Context, userID, noteID uint64) error { return nil }
func (r *nu_fakeRepo) RemoveFavorite(ctx context.Context, userID, noteID uint64) error { return nil }

// errRepo simulates repository failures for all methods
type errRepo struct{}

func (r *errRepo) GetNotes(ctx context.Context, userID uint64) ([]models.Note, error) { return nil, fmt.Errorf("db") }
func (r *errRepo) CreateNote(ctx context.Context, userID uint64) (*models.Note, error) { return nil, fmt.Errorf("db") }
func (r *errRepo) GetNoteById(ctx context.Context, noteID uint64, userID uint64) (*models.Note, error) { return nil, fmt.Errorf("db") }
func (r *errRepo) UpdateNote(ctx context.Context, noteID uint64, title *string, isArchived *bool) (*models.Note, error) { return nil, fmt.Errorf("db") }
func (r *errRepo) DeleteNote(ctx context.Context, noteID uint64) error { return fmt.Errorf("db") }
func (r *errRepo) AddFavorite(ctx context.Context, userID, noteID uint64) error { return fmt.Errorf("db") }
func (r *errRepo) RemoveFavorite(ctx context.Context, userID, noteID uint64) error { return fmt.Errorf("db") }

func TestNotesUsecase_basicFlows(t *testing.T) {
	repo := &nu_fakeRepo{owner:1}
	u := NewNotesUsecase(repo)

	notes, err := u.GetAllNotes(context.Background(), 1)
	assert.NoError(t, err)
	if assert.Len(t, notes, 1) {
		assert.Equal(t, uint64(1), notes[0].ID)
	}

	n, err := u.CreateNote(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), n.ID)

	got, err := u.GetNoteById(context.Background(), 1, 1)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), got.ID)

	title := "NewTitle"
	updated, err := u.UpdateNote(context.Background(), 1, 1, &title, nil)
	assert.NoError(t, err)
	assert.Equal(t, "NewTitle", updated.Title)

	err = u.DeleteNote(context.Background(), 1, 1)
	assert.NoError(t, err)

	err = u.AddFavorite(context.Background(), 1, 1)
	assert.NoError(t, err)

	err = u.RemoveFavorite(context.Background(), 1, 1)
	assert.NoError(t, err)
}

func TestNotesUsecase_AccessDeniedAndErrors(t *testing.T) {
	// repo returns a note owned by someone else -> access denied
	repo := &nu_fakeRepo{owner:2}
	u := NewNotesUsecase(repo)

	_, err := u.GetNoteById(context.Background(), 1, 1)
	assert.Error(t, err)

	// use package-scope errRepo which simulates repository failures
	er := &errRepo{}
	u2 := NewNotesUsecase(er)
	_, err = u2.GetAllNotes(context.Background(), 1)
	assert.Error(t, err)
}

func TestNotesUsecase_UpdateDelete_ErrorFlows(t *testing.T) {
	// errRepo returns errors for all operations
	er := &errRepo{}
	u := NewNotesUsecase(er)

	title := "T"
	_, err := u.UpdateNote(context.Background(), 1, 1, &title, nil)
	assert.Error(t, err)

	err = u.DeleteNote(context.Background(), 1, 1)
	assert.Error(t, err)

	err = u.AddFavorite(context.Background(), 1, 1)
	assert.Error(t, err)

	err = u.RemoveFavorite(context.Background(), 1, 1)
	assert.Error(t, err)
}
