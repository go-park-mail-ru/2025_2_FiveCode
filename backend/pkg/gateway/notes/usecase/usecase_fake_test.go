package usecase

import (
	"backend/models"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeNotesRepo2 struct{
    notes []models.Note
}
func (f *fakeNotesRepo2) GetNotes(ctx context.Context, userID uint64) ([]models.Note, error) { return f.notes, nil }
func (f *fakeNotesRepo2) CreateNote(ctx context.Context, userID uint64) (*models.Note, error) { return &models.Note{ID:1, OwnerID:userID}, nil }
func (f *fakeNotesRepo2) GetNoteById(ctx context.Context, noteID uint64, userID uint64) (*models.Note, error) { return &models.Note{ID:noteID, OwnerID:userID}, nil }
func (f *fakeNotesRepo2) UpdateNote(ctx context.Context, noteID uint64, title *string, isArchived *bool) (*models.Note, error) { return &models.Note{ID:noteID, OwnerID:1, Title: "ok"}, nil }
func (f *fakeNotesRepo2) DeleteNote(ctx context.Context, noteID uint64) error { return nil }
func (f *fakeNotesRepo2) AddFavorite(ctx context.Context, userID, noteID uint64) error { return nil }
func (f *fakeNotesRepo2) RemoveFavorite(ctx context.Context, userID, noteID uint64) error { return nil }

func TestNotesUsecase_basicPaths(t *testing.T) {
    repo := &fakeNotesRepo2{notes: []models.Note{{ID:1, OwnerID:1}}}
    uc := NewNotesUsecase(repo)
    ctx := context.Background()

    notes, err := uc.GetAllNotes(ctx, 1)
    assert.NoError(t, err)
    assert.NotNil(t, notes)

    n, err := uc.CreateNote(ctx, 2)
    assert.NoError(t, err)
    assert.Equal(t, uint64(2), n.OwnerID)

    got, err := uc.GetNoteById(ctx, 1, 1)
    assert.NoError(t, err)
    assert.Equal(t, uint64(1), got.ID)

    _, err = uc.UpdateNote(ctx, 1, 1, nil, nil)
    assert.NoError(t, err)

    err = uc.DeleteNote(ctx, 1, 1)
    assert.NoError(t, err)

    // favorites rely on checkNoteAccess; fake repo returns owner=1
    err = uc.AddFavorite(ctx, 1, 1)
    assert.NoError(t, err)
    err = uc.RemoveFavorite(ctx, 1, 1)
    assert.NoError(t, err)
}
