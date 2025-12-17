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
	Repository        NotesRepository
	SharingRepository SharingRepository
}

type NotesRepository interface {
	GetNotes(ctx context.Context, userID uint64) ([]models.Note, error)
	CreateNote(ctx context.Context, userID uint64, parentNoteID *uint64) (*models.Note, error)
	GetNoteById(ctx context.Context, noteID uint64, userID uint64) (*models.Note, error)
	GetNoteByShareUUID(ctx context.Context, shareUUID string) (*models.Note, error)
	UpdateNote(ctx context.Context, noteID uint64, title *string, isArchived *bool) (*models.Note, error)
	DeleteNote(ctx context.Context, noteID uint64) error
	AddFavorite(ctx context.Context, userID, noteID uint64) error
	RemoveFavorite(ctx context.Context, userID, noteID uint64) error
	SearchNotes(ctx context.Context, userID uint64, query string) (*models.SearchNotesResponse, error)
	SetIcon(ctx context.Context, noteID, iconFileID uint64) error
	SetHeader(ctx context.Context, noteID, headerFileID uint64) error
}

type SharingRepository interface {
	CheckNoteAccess(ctx context.Context, noteID, userID uint64) (*models.NoteAccessInfo, error)
}

func NewNoteUsecase(repository NotesRepository, sharingRepository SharingRepository) *NoteUsecase {
	return &NoteUsecase{
		Repository:        repository,
		SharingRepository: sharingRepository,
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

func (u *NoteUsecase) CreateNote(ctx context.Context, userID uint64, parentNoteID *uint64) (*models.Note, error) {
	log := logger.FromContext(ctx)

	if parentNoteID != nil && *parentNoteID > 0 {

		accessInfo, err := u.SharingRepository.CheckNoteAccess(ctx, *parentNoteID, userID)
		if err != nil {
			log.Error().Err(err).Uint64("parent_note_id", *parentNoteID).Msg("failed to check access")
			return nil, fmt.Errorf("failed to check access: %w", err)
		}

		if !accessInfo.HasAccess {
			log.Warn().Uint64("user_id", userID).Uint64("parent_note_id", *parentNoteID).Msg("no access to parent note")
			return nil, fmt.Errorf("no access to parent note")
		}

		if !accessInfo.CanEdit {
			log.Warn().Uint64("user_id", userID).Uint64("parent_note_id", *parentNoteID).Str("role", string(accessInfo.Role)).Msg("user cannot create sub-note: editor rights required")
			return nil, fmt.Errorf("no access to parent note")
		}

		parentNote, err := u.Repository.GetNoteById(ctx, *parentNoteID, userID)
		if err != nil {
			log.Error().Err(err).Uint64("parent_note_id", *parentNoteID).Msg("parent note not found")
			return nil, fmt.Errorf("parent note not found: %w", err)
		}

		if parentNote.ParentNoteID != nil {
			log.Warn().Uint64("parent_note_id", *parentNoteID).Msg("attempt to create sub-note of a sub-note")
			return nil, fmt.Errorf("cannot create sub-note of a sub-note: maximum nesting level is 1")
		}
	}

	note, err := u.Repository.CreateNote(ctx, userID, parentNoteID)
	if err != nil {
		log.Error().Err(err).Msg("failed to create note in repository")
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	return note, nil
}

func (u *NoteUsecase) GetNoteById(ctx context.Context, userID, noteID uint64) (*models.Note, error) {
	log := logger.FromContext(ctx)

	accessInfo, err := u.SharingRepository.CheckNoteAccess(ctx, noteID, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to check note access")
		return nil, fmt.Errorf("failed to check note access: %w", err)
	}

	if !accessInfo.HasAccess {
		log.Warn().Uint64("user_id", userID).Uint64("note_id", noteID).Msg("user has no access to note")
		return nil, constants.ErrNoAccess
	}

	note, err := u.Repository.GetNoteById(ctx, noteID, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get note by id from repository")
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	return note, nil
}

func (u *NoteUsecase) UpdateNote(ctx context.Context, userID uint64, noteID uint64, title *string, isArchived *bool) (*models.Note, error) {
	log := logger.FromContext(ctx)

	accessInfo, err := u.SharingRepository.CheckNoteAccess(ctx, noteID, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to check note access for update")
		return nil, fmt.Errorf("failed to check note access: %w", err)
	}

	if !accessInfo.HasAccess {
		log.Warn().Uint64("user_id", userID).Uint64("note_id", noteID).Msg("user has no access to note")
		return nil, constants.ErrNoAccess
	}

	if !accessInfo.CanEdit {
		log.Warn().Uint64("user_id", userID).Uint64("note_id", noteID).Str("role", string(accessInfo.Role)).Msg("user cannot edit note")
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

	accessInfo, err := u.SharingRepository.CheckNoteAccess(ctx, noteID, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to check note access for deletion")
		return fmt.Errorf("failed to check note access: %w", err)
	}

	if !accessInfo.HasAccess {
		log.Warn().Uint64("user_id", userID).Uint64("note_id", noteID).Msg("user has no access to note")
		return constants.ErrNoAccess
	}

	if !accessInfo.IsOwner {
		log.Warn().Uint64("user_id", userID).Uint64("note_id", noteID).Msg("only owner can delete note")
		return constants.ErrNoAccess
	}

	if err := u.Repository.DeleteNote(ctx, noteID); err != nil {
		log.Error().Err(err).Msg("failed to delete note in repository")
		return fmt.Errorf("failed to delete note: %w", err)
	}

	return nil
}

func (u *NoteUsecase) AddFavorite(ctx context.Context, userID, noteID uint64) error {
	log := logger.FromContext(ctx)

	accessInfo, err := u.SharingRepository.CheckNoteAccess(ctx, noteID, userID)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("failed to check note access")
		return fmt.Errorf("failed to check note access: %w", err)
	}

	if !accessInfo.HasAccess {
		log.Warn().Uint64("user_id", userID).Uint64("note_id", noteID).Msg("user has no access to note")
		return constants.ErrNoAccess
	}

	return u.Repository.AddFavorite(ctx, userID, noteID)
}

func (u *NoteUsecase) RemoveFavorite(ctx context.Context, userID, noteID uint64) error {
	log := logger.FromContext(ctx)

	accessInfo, err := u.SharingRepository.CheckNoteAccess(ctx, noteID, userID)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("failed to check note access")
		return fmt.Errorf("failed to check note access: %w", err)
	}

	if !accessInfo.HasAccess {
		log.Warn().Uint64("user_id", userID).Uint64("note_id", noteID).Msg("user has no access to note")
		return constants.ErrNoAccess
	}

	return u.Repository.RemoveFavorite(ctx, userID, noteID)
}

func (u *NoteUsecase) GetNoteByShareUUID(ctx context.Context, shareUUID string) (*models.Note, error) {
	log := logger.FromContext(ctx)

	note, err := u.Repository.GetNoteByShareUUID(ctx, shareUUID)
	if err != nil {
		log.Error().Err(err).Str("share_uuid", shareUUID).Msg("failed to get note by share_uuid")
		return nil, fmt.Errorf("failed to get note by share_uuid: %w", err)
	}

	return note, nil
}

func (u *NoteUsecase) SearchNotes(ctx context.Context, userID uint64, query string) (*models.SearchNotesResponse, error) {
	log := logger.FromContext(ctx)

	if query == "" {
		log.Warn().Msg("search query is empty")
		return nil, fmt.Errorf("search query cannot be empty")
	}

	if len(query) > 200 {
		log.Warn().Int("length", len(query)).Msg("search query too long")
		return nil, fmt.Errorf("search query too long (max 200 characters)")
	}

	searchResult, err := u.Repository.SearchNotes(ctx, userID, query)
	if err != nil {
		log.Error().Err(err).Msg("failed to search notes in repository")
		return nil, fmt.Errorf("failed to search notes in repository: %w", err)
	}

	return searchResult, nil
}

func (u *NoteUsecase) SetIcon(ctx context.Context, userID, noteID, iconFileID uint64) (*models.Note, error) {
	log := logger.FromContext(ctx)

	accessInfo, err := u.SharingRepository.CheckNoteAccess(ctx, noteID, userID)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Uint64("user_id", userID).Msg("failed to check note access")
		return nil, fmt.Errorf("failed to check note access: %w", err)
	}

	if !accessInfo.HasAccess {
		log.Warn().Uint64("user_id", userID).Uint64("note_id", noteID).Msg("user has no access to note")
		return nil, constants.ErrNoAccess
	}

	if !accessInfo.CanEdit {
		log.Warn().Uint64("user_id", userID).Uint64("note_id", noteID).Msg("user cannot edit note")
		return nil, constants.ErrNoAccess
	}

	if err := u.Repository.SetIcon(ctx, noteID, iconFileID); err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("failed to set icon in repository")
		return nil, fmt.Errorf("failed to set icon: %w", err)
	}

	note, err := u.Repository.GetNoteById(ctx, noteID, userID)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("failed to get note after setting icon")
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	return note, nil
}

func (u *NoteUsecase) SetHeader(ctx context.Context, userID, noteID, headerFileID uint64) (*models.Note, error) {
	log := logger.FromContext(ctx)

	accessInfo, err := u.SharingRepository.CheckNoteAccess(ctx, noteID, userID)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Uint64("user_id", userID).Msg("failed to check note access")
		return nil, fmt.Errorf("failed to check note access: %w", err)
	}

	if !accessInfo.HasAccess {
		log.Warn().Uint64("user_id", userID).Uint64("note_id", noteID).Msg("user has no access to note")
		return nil, constants.ErrNoAccess
	}

	if !accessInfo.CanEdit {
		log.Warn().Uint64("user_id", userID).Uint64("note_id", noteID).Msg("user cannot edit note")
		return nil, constants.ErrNoAccess
	}

	if err := u.Repository.SetHeader(ctx, noteID, headerFileID); err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("failed to set header in repository")
		return nil, fmt.Errorf("failed to set header: %w", err)
	}

	note, err := u.Repository.GetNoteById(ctx, noteID, userID)
	if err != nil {
		log.Error().Err(err).Uint64("note_id", noteID).Msg("failed to get note after setting header")
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	return note, nil
}
