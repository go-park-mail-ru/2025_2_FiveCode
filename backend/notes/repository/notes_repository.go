package notesRepository

import (
	"backend/models"
	"backend/store"
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

func (r *NotesRepository) GetNotes(userID uint64) ([]models.Note, error) {
	notes := r.Store.ListNotes(userID)
	return notes, nil
}

func (r *NotesRepository) CreateNote(userID uint64) (*models.Note, error) {
	note, err := r.Store.CreateNote(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	return note, nil
}

func (r *NotesRepository) GetNoteById(noteID uint64) (*models.Note, error) {
	note, err := r.Store.GetNoteById(noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	return note, nil
}
