package notesRepository

import (
	"backend/models"
	"backend/store"
)

type NotesRepository struct {
	Store *store.Store
}

func NewNotesRepository(store *store.Store) *NotesRepository {
	return &NotesRepository{
		Store: store,
	}
}

func (r *NotesRepository) GetNotes(ownerID uint64) ([]models.Note, error) {
	notes := r.Store.ListNotes(ownerID)
	return notes, nil
}
