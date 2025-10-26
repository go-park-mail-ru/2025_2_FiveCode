package notesUsecase

import (
	"backend/models"
	namederrors "backend/named_errors"
	"fmt"
)

type NotesUsecase struct {
	Repository NotesRepository
}

type NotesRepository interface {
	GetNotes(userID uint64) ([]models.Note, error)
	CreateNote(userID uint64) (*models.Note, error)
	GetNoteById(noteID uint64) (*models.Note, error)
	UpdateNote(noteID uint64, title *string, isArchived *bool) (*models.Note, error)
	DeleteNote(noteID uint64) error
}

func NewNotesUsecase(Repository NotesRepository) *NotesUsecase {
	return &NotesUsecase{
		Repository: Repository,
	}
}

func (u *NotesUsecase) GetAllNotes(userID uint64) ([]models.Note, error) {
	notes, err := u.Repository.GetNotes(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notes: %w", err)
	}
	return notes, nil
}

func (u *NotesUsecase) CreateNote(userID uint64) (*models.Note, error) {
	note, err := u.Repository.CreateNote(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	return note, nil
}

func (u *NotesUsecase) GetNoteById(userID, noteID uint64) (*models.Note, error) {
	note, err := u.Repository.GetNoteById(noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	if note.OwnerID != userID {
		return nil, namederrors.ErrNoAccess
	}

	return note, nil
}

func (u *NotesUsecase) UpdateNote(userID uint64, noteID uint64, title *string, isArchived *bool) (*models.Note, error) {
	note, err := u.Repository.GetNoteById(noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	if note.OwnerID != userID {
		return nil, namederrors.ErrNoAccess
	}

	updatedNote, err := u.Repository.UpdateNote(noteID, title, isArchived)
	if err != nil {
		return nil, err
	}

	return updatedNote, nil
}

func (u *NotesUsecase) DeleteNote(userID uint64, noteID uint64) error {
	note, err := u.Repository.GetNoteById(noteID)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}

	if note.OwnerID != userID {
		return namederrors.ErrNoAccess
	}

	err = u.Repository.DeleteNote(noteID)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	return nil
}
