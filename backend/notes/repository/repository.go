package Repository

import (
	"backend/models"
	"backend/store"
	"context"
	"errors"
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
		return nil, errors.New("failed to create note: " + err.Error())
	}

	return note, nil
}

func (r *NotesRepository) GetNoteById(ctx context.Context, noteID uint64) (*models.Note, error) {
	note, err := r.Store.GetNoteById(noteID)
	if err != nil {
		return nil, errors.New("failed to get note: " + err.Error())
	}

	return note, nil
}

func (r *NotesRepository) UpdateNote(ctx context.Context, noteID uint64, title *string, isArchived *bool) (*models.Note, error) {
	note, err := r.Store.UpdateNote(noteID, title, isArchived)
	if err != nil {
		return nil, errors.New("failed to update note: " + err.Error())
	}

	return note, nil
}

func (r *NotesRepository) DeleteNote(ctx context.Context, noteID uint64) error {
	err := r.Store.DeleteNote(noteID)
	if err != nil {
		return errors.New("failed to delete note: " + err.Error())
	}

	return nil
}
