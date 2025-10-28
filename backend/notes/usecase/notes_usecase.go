package notesUsecase

import (
	"backend/models"
	namederrors "backend/named_errors"
	"context"
	"fmt"
)

type NotesUsecase struct {
	Repository NotesRepository
}

type NotesRepository interface {
	GetNotes(ctx context.Context, userID uint64) ([]models.Note, error)
	CreateNote(ctx context.Context, userID uint64) (*models.Note, error)
	GetNoteById(ctx context.Context, noteID uint64) (*models.Note, error)
	UpdateNote(ctx context.Context, noteID uint64, title *string, isArchived *bool) (*models.Note, error)
	DeleteNote(ctx context.Context, noteID uint64) error
}

func NewNotesUsecase(Repository NotesRepository) *NotesUsecase {
	return &NotesUsecase{
		Repository: Repository,
	}
}

func (u *NotesUsecase) GetAllNotes(ctx context.Context, userID uint64) ([]models.Note, error) {
	notes, err := u.Repository.GetNotes(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notes: %w", err)
	}
	return notes, nil
}

func (u *NotesUsecase) CreateNote(ctx context.Context, userID uint64) (*models.Note, error) {
	note, err := u.Repository.CreateNote(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	return note, nil
}

func (u *NotesUsecase) GetNoteById(ctx context.Context, userID, noteID uint64) (*models.Note, error) {
	note, err := u.Repository.GetNoteById(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	if note.OwnerID != userID {
		return nil, namederrors.ErrNoAccess
	}

	return note, nil
}

func (u *NotesUsecase) UpdateNote(ctx context.Context, userID uint64, noteID uint64, title *string, isArchived *bool) (*models.Note, error) {
	note, err := u.Repository.GetNoteById(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	if note.OwnerID != userID {
		return nil, namederrors.ErrNoAccess
	}

	updatedNote, err := u.Repository.UpdateNote(ctx, noteID, title, isArchived)
	if err != nil {
		return nil, err
	}

	return updatedNote, nil
}

func (u *NotesUsecase) DeleteNote(ctx context.Context, userID uint64, noteID uint64) error {
	note, err := u.Repository.GetNoteById(ctx, noteID)
	if err != nil {
		return fmt.Errorf("failed to get note: %w", err)
	}

	if note.OwnerID != userID {
		return namederrors.ErrNoAccess
	}

	err = u.Repository.DeleteNote(ctx, noteID)
	if err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	return nil
}
