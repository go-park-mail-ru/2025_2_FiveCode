package notesRepository

import (
	"backend/models"
	"backend/store"
	"context"
	"fmt"
)

type NotesRepository struct {
	Store *store.Store
}

func NewNotesRepository(store *store.Store) *NotesRepository {
	return &NotesRepository{
		Store: store,
	}
}

func (r *NotesRepository) GetNotes(ctx context.Context, userID uint64) ([]models.Note, error) {
	notes := r.Store.ListNotes(userID)
	return notes, nil
}

func (r *NotesRepository) CreateNote(ctx context.Context, userID uint64) (*models.Note, error) {
	note, err := r.Store.CreateNote(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	return note, nil
}

func (r *NotesRepository) GetNoteById(ctx context.Context, noteID uint64) (*models.Note, error) {
	note, err := r.Store.GetNoteById(noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	return note, nil
}

func (r *NotesRepository) UpdateNote(ctx context.Context, noteID uint64, title *string, isArchived *bool) (*models.Note, error) {
	note, err := r.Store.UpdateNote(noteID, title, isArchived)
	if err != nil {
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	return note, nil
}

func (r *NotesRepository) DeleteNote(ctx context.Context, noteID uint64) error {
	err := r.Store.DeleteNote(noteID)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	return nil
}
