package usecase

import (
	"backend/notes_service/internal/constants"
	"backend/notes_service/internal/models"
	"backend/notes_service/logger"
	"context"
	"fmt"
)

//go:generate mockgen -source=usecase.go -destination=../mock/mock_usecase.go -package=mock
type NoteUsecase struct {
	Repository NotesRepository
}

type NotesRepository interface {
	GetNotes(ctx context.Context, userID uint64) ([]models.Note, error)
	CreateNote(ctx context.Context, userID uint64) (*models.Note, error)
	GetNoteById(ctx context.Context, noteID uint64, userID uint64) (*models.Note, error)
	UpdateNote(ctx context.Context, noteID uint64, title *string, isArchived *bool) (*models.Note, error)
	DeleteNote(ctx context.Context, noteID uint64) error
	AddFavorite(ctx context.Context, userID, noteID uint64) error
	RemoveFavorite(ctx context.Context, userID, noteID uint64) error
}

func NewNoteUsecase(Repository NotesRepository) *NoteUsecase {
	return &NoteUsecase{
		Repository: Repository,
	}
}

func (u *NoteUsecase) GetAllNotes(ctx context.Context, userID uint64) ([]models.Note, error) {
	log := logger.FromContext(ctx)
	notes, err := u.Repository.GetNotes(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get notes from repository")
		return nil, fmt.Errorf("failed to get notes: %w", err)
	}
	return notes, nil
}

func (u *NoteUsecase) CreateNote(ctx context.Context, userID uint64) (*models.Note, error) {
	log := logger.FromContext(ctx)
	note, err := u.Repository.CreateNote(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to create note in repository")
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	return note, nil
}

func (u *NoteUsecase) GetNoteById(ctx context.Context, userID, noteID uint64) (*models.Note, error) {
	log := logger.FromContext(ctx)
	note, err := u.Repository.GetNoteById(ctx, noteID, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get note by id from repository")
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	if note.OwnerID != userID {
		log.Warn().Uint64("user_id", userID).Uint64("owner_id", note.OwnerID).Msg("user access denied")
		return nil, constants.ErrNoAccess
	}

	return note, nil
}

func (u *NoteUsecase) UpdateNote(ctx context.Context, userID uint64, noteID uint64, title *string, isArchived *bool) (*models.Note, error) {
	log := logger.FromContext(ctx)
	note, err := u.Repository.GetNoteById(ctx, noteID, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get note for update")
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	if note.OwnerID != userID {
		log.Warn().Uint64("user_id", userID).Uint64("owner_id", note.OwnerID).Msg("user access denied for update")
		return nil, constants.ErrNoAccess
	}

	updatedNote, err := u.Repository.UpdateNote(ctx, noteID, title, isArchived)
	if err != nil {
		log.Error().Err(err).Msg("failed to update note in repository")
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	return updatedNote, nil
}

func (u *NoteUsecase) DeleteNote(ctx context.Context, userID uint64, noteID uint64) error {
	log := logger.FromContext(ctx)
	note, err := u.Repository.GetNoteById(ctx, noteID, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get note for deletion")
		return fmt.Errorf("failed to get note: %w", err)
	}

	if note.OwnerID != userID {
		log.Warn().Uint64("user_id", userID).Uint64("owner_id", note.OwnerID).Msg("user access denied for deletion")
		return constants.ErrNoAccess
	}

	if err := u.Repository.DeleteNote(ctx, noteID); err != nil {
		log.Error().Err(err).Msg("failed to delete note in repository")
		return fmt.Errorf("failed to delete note: %w", err)
	}

	return nil
}

func (u *NoteUsecase) AddFavorite(ctx context.Context, userID, noteID uint64) error {
	if err := u.checkNoteAccess(ctx, userID, noteID); err != nil {
		return err
	}
	return u.Repository.AddFavorite(ctx, userID, noteID)
}

func (u *NoteUsecase) RemoveFavorite(ctx context.Context, userID, noteID uint64) error {
	if err := u.checkNoteAccess(ctx, userID, noteID); err != nil {
		return err
	}
	return u.Repository.RemoveFavorite(ctx, userID, noteID)
}

func (u *NoteUsecase) checkNoteAccess(ctx context.Context, userID, noteID uint64) error {
	log := logger.FromContext(ctx)
	note, err := u.Repository.GetNoteById(ctx, noteID, userID)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("failed to get note for access check")
		return fmt.Errorf("failed to get note by id: %w", err)
	}

	if note.OwnerID != userID {
		log.Warn().Uint64("user_id", userID).Uint64("note_id", noteID).Uint64("owner_id", note.OwnerID).Msg("user access denied to note")
		return constants.ErrNoAccess
	}

	return nil
}
