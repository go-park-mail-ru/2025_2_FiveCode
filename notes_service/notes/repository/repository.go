package repository

import (
	"backend/notes_service/internal/constants"
	"backend/notes_service/internal/models"
	"backend/notes_service/logger"
	"backend/pkg/store"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type NotesRepository struct {
	db store.DB
}

func NewNotesRepository(db store.DB) *NotesRepository {
	return &NotesRepository{
		db: db,
	}
}

func (r *NotesRepository) CreateNote(ctx context.Context, userID uint64, parentNoteID *uint64) (*models.Note, error) {
	log := logger.FromContext(ctx)
	now := time.Now().UTC()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to begin transaction")
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Error().Err(err).Msg("failed to rollback transaction")
		}
	}()

	shareUUID := uuid.New().String()

	noteQuery := `
        INSERT INTO note (owner_id, parent_note_id, title, is_archived, is_shared, share_uuid, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, owner_id, parent_note_id, title, icon_file_id, 
                  is_archived, is_shared, share_uuid, created_at, updated_at, deleted_at
    `
	defaultTitle := "Новая заметка"

	note := &models.Note{}
	var parentNoteIDResult, iconFileID sql.NullInt64
	var shareUUIDResult sql.NullString
	var deletedAt sql.NullTime

	var parentNoteIDParam interface{}
	if parentNoteID != nil && *parentNoteID > 0 {
		parentNoteIDParam = *parentNoteID
	} else {
		parentNoteIDParam = nil
	}

	err = tx.QueryRowContext(ctx, noteQuery, userID, parentNoteIDParam, defaultTitle, false, false, shareUUID, now, now).Scan(
		&note.ID,
		&note.OwnerID,
		&parentNoteIDResult,
		&note.Title,
		&iconFileID,
		&note.IsArchived,
		&note.IsShared,
		&shareUUIDResult,
		&note.CreatedAt,
		&note.UpdatedAt,
		&deletedAt,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to create note entry")
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	blockQuery := `
        INSERT INTO block (note_id, type, position, last_edited_by, created_at, updated_at) 
        VALUES ($1, 'text', 1.0, $2, $3, $4)
        RETURNING id
    `
	var blockID uint64
	err = tx.QueryRowContext(ctx, blockQuery, note.ID, userID, now, now).Scan(&blockID)
	if err != nil {
		log.Error().Err(err).Msg("failed to create initial block entry")
		return nil, fmt.Errorf("failed to create initial block: %w", err)
	}

	textQuery := `
        INSERT INTO block_text (block_id, text, created_at, updated_at) VALUES ($1, '', $2, $3)
    `
	_, err = tx.ExecContext(ctx, textQuery, blockID, now, now)
	if err != nil {
		log.Error().Err(err).Msg("failed to create initial block_text entry")
		return nil, fmt.Errorf("failed to create initial block_text: %w", err)
	}

	if err = tx.Commit(); err != nil {
		log.Error().Err(err).Msg("failed to commit transaction")
		return nil, fmt.Errorf("failed to commit create note transaction: %w", err)
	}

	if parentNoteIDResult.Valid {
		val := uint64(parentNoteIDResult.Int64)
		note.ParentNoteID = &val
	}
	if iconFileID.Valid {
		val := uint64(iconFileID.Int64)
		note.IconFileID = &val
	}
	if shareUUIDResult.Valid {
		note.ShareUUID = &shareUUIDResult.String
	}
	if deletedAt.Valid {
		note.DeletedAt = &deletedAt.Time
	}

	return note, nil
}

func (r *NotesRepository) GetNotes(ctx context.Context, userID uint64) ([]models.Note, error) {
	log := logger.FromContext(ctx)

	query := `
        WITH accessible_notes AS (
            SELECT DISTINCT n.id
            FROM note n
            LEFT JOIN note_permission np ON n.id = np.note_id AND np.granted_to = $1
            WHERE (n.owner_id = $1 OR np.note_permission_id IS NOT NULL)
              AND n.deleted_at IS NULL
        )
        SELECT DISTINCT n.id, n.owner_id, n.parent_note_id, n.title, n.icon_file_id,
               n.is_archived, n.is_shared, n.share_uuid, n.created_at, n.updated_at,
               f.user_id IS NOT NULL AS is_favorite
        FROM note n
        LEFT JOIN favorite f ON n.id = f.note_id AND f.user_id = $1
        WHERE (
            n.id IN (SELECT id FROM accessible_notes)  
            OR 
            n.parent_note_id IN (SELECT id FROM accessible_notes) 
        )
        AND n.deleted_at IS NULL
        ORDER BY n.updated_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list notes")
		return nil, fmt.Errorf("failed to list notes: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close rows")
		}
	}()

	notes := make([]models.Note, 0)

	for rows.Next() {
		var note models.Note
		var parentNoteID, iconFileID sql.NullInt64
		var shareUUID sql.NullString

		err := rows.Scan(
			&note.ID,
			&note.OwnerID,
			&parentNoteID,
			&note.Title,
			&iconFileID,
			&note.IsArchived,
			&note.IsShared,
			&shareUUID,
			&note.CreatedAt,
			&note.UpdatedAt,
			&note.IsFavorite,
		)
		if err != nil {
			log.Error().Err(err).Msg("failed to scan note")
			return nil, fmt.Errorf("failed to scan note: %w", err)
		}

		if parentNoteID.Valid {
			val := uint64(parentNoteID.Int64)
			note.ParentNoteID = &val
		}
		if iconFileID.Valid {
			val := uint64(iconFileID.Int64)
			note.IconFileID = &val
		}
		if shareUUID.Valid {
			note.ShareUUID = &shareUUID.String
		}

		notes = append(notes, note)
	}

	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("error iterating notes")
		return nil, fmt.Errorf("error iterating notes: %w", err)
	}

	return notes, nil
}

func (r *NotesRepository) GetNoteById(ctx context.Context, noteID uint64, userID uint64) (*models.Note, error) {
	log := logger.FromContext(ctx)

	query := `
		SELECT n.id, n.owner_id, n.parent_note_id, n.title, n.icon_file_id,
		       n.is_archived, n.is_shared, n.share_uuid, n.created_at, n.updated_at, n.deleted_at,
		       f.user_id IS NOT NULL AS is_favorite
		FROM note n
		LEFT JOIN favorite f ON n.id = f.note_id AND f.user_id = $2
		WHERE n.id = $1 AND n.deleted_at IS NULL
	`

	note := &models.Note{}
	var parentNoteID, iconFileID sql.NullInt64
	var shareUUID sql.NullString
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, noteID, userID).Scan(
		&note.ID,
		&note.OwnerID,
		&parentNoteID,
		&note.Title,
		&iconFileID,
		&note.IsArchived,
		&note.IsShared,
		&shareUUID,
		&note.CreatedAt,
		&note.UpdatedAt,
		&deletedAt,
		&note.IsFavorite,
	)

	if errors.Is(err, sql.ErrNoRows) {
		log.Warn().Err(err).Uint64("note_id", noteID).Msg("note not found")
		return nil, constants.ErrNotFound
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to get note")
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	if parentNoteID.Valid {
		val := uint64(parentNoteID.Int64)
		note.ParentNoteID = &val
	}
	if iconFileID.Valid {
		val := uint64(iconFileID.Int64)
		note.IconFileID = &val
	}
	if shareUUID.Valid {
		note.ShareUUID = &shareUUID.String
	}

	return note, nil
}

func (r *NotesRepository) UpdateNote(ctx context.Context, noteID uint64, title *string, isArchived *bool) (*models.Note, error) {
	log := logger.FromContext(ctx)

	checkQuery := `SELECT 1 FROM note WHERE id = $1 AND deleted_at IS NULL`
	var exists int
	err := r.db.QueryRowContext(ctx, checkQuery, noteID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		log.Warn().Err(err).Uint64("note_id", noteID).Msg("note not found for update")
		return nil, constants.ErrNotFound
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to check note existence")
		return nil, fmt.Errorf("failed to check note existence: %w", err)
	}

	query := `UPDATE note SET updated_at = $1`
	args := []interface{}{time.Now().UTC()}
	argIndex := 2

	if title != nil {
		query += fmt.Sprintf(", title = $%d", argIndex)
		args = append(args, *title)
		argIndex++
	}

	if isArchived != nil {
		query += fmt.Sprintf(", is_archived = $%d", argIndex)
		args = append(args, *isArchived)
		argIndex++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIndex)
	args = append(args, noteID)

	query += ` RETURNING id, owner_id, parent_note_id, title, icon_file_id,
	          is_archived, is_shared, share_uuid, created_at, updated_at`

	note := &models.Note{}
	var parentNoteID, iconFileID sql.NullInt64
	var shareUUID sql.NullString

	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&note.ID,
		&note.OwnerID,
		&parentNoteID,
		&note.Title,
		&iconFileID,
		&note.IsArchived,
		&note.IsShared,
		&shareUUID,
		&note.CreatedAt,
		&note.UpdatedAt,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to update note")
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	if parentNoteID.Valid {
		val := uint64(parentNoteID.Int64)
		note.ParentNoteID = &val
	}
	if iconFileID.Valid {
		val := uint64(iconFileID.Int64)
		note.IconFileID = &val
	}
	if shareUUID.Valid {
		note.ShareUUID = &shareUUID.String
	}

	return note, nil
}

func (r *NotesRepository) DeleteNote(ctx context.Context, noteID uint64) error {
	log := logger.FromContext(ctx)

	query := `DELETE FROM note WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, noteID)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete note")
		return fmt.Errorf("failed to delete note: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("failed to get rows affected")
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		log.Warn().Uint64("note_id", noteID).Msg("note not found for deletion")
		return constants.ErrNotFound
	}

	return nil
}

func (r *NotesRepository) AddFavorite(ctx context.Context, userID, noteID uint64) error {
	log := logger.FromContext(ctx)
	now := time.Now().UTC()

	query := `INSERT INTO favorite (user_id, note_id, created_at, updated_at) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, userID, noteID, now, now)
	if err != nil {
		log.Error().Err(err).Msg("failed to add favorite")
		return fmt.Errorf("failed to add favorite: %w", err)
	}
	return nil
}

func (r *NotesRepository) RemoveFavorite(ctx context.Context, userID, noteID uint64) error {
	log := logger.FromContext(ctx)
	query := `DELETE FROM favorite WHERE user_id = $1 AND note_id = $2`
	_, err := r.db.ExecContext(ctx, query, userID, noteID)
	if err != nil {
		log.Error().Err(err).Msg("failed to remove favorite")
		return fmt.Errorf("failed to remove favorite: %w", err)
	}
	return nil
}

func (r *NotesRepository) CheckNoteOwnership(ctx context.Context, noteID uint64, userID uint64) (bool, error) {
	log := logger.FromContext(ctx)

	query := `SELECT owner_id FROM note WHERE id = $1 AND deleted_at IS NULL`

	var ownerID uint64
	err := r.db.QueryRowContext(ctx, query, noteID).Scan(&ownerID)
	if errors.Is(err, sql.ErrNoRows) {
		log.Warn().Uint64("note_id", noteID).Msg("note not found")
		return false, constants.ErrNotFound
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to check note ownership")
		return false, fmt.Errorf("failed to check note ownership: %w", err)
	}

	return ownerID == userID, nil
}

func (r *NotesRepository) GetNoteByShareUUID(ctx context.Context, shareUUID string) (*models.Note, error) {
	log := logger.FromContext(ctx)

	query := `
		SELECT n.id, n.owner_id, n.parent_note_id, n.title, n.icon_file_id,
		       n.is_archived, n.is_shared, n.share_uuid, n.public_access_level,
		       n.created_at, n.updated_at, n.deleted_at
		FROM note n
		WHERE n.share_uuid = $1 AND n.deleted_at IS NULL
	`

	note := &models.Note{}
	var parentNoteID, iconFileID sql.NullInt64
	var shareUUIDResult sql.NullString
	var publicAccessLevel sql.NullString
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, shareUUID).Scan(
		&note.ID,
		&note.OwnerID,
		&parentNoteID,
		&note.Title,
		&iconFileID,
		&note.IsArchived,
		&note.IsShared,
		&shareUUIDResult,
		&publicAccessLevel,
		&note.CreatedAt,
		&note.UpdatedAt,
		&deletedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		log.Warn().Str("share_uuid", shareUUID).Msg("note not found by share_uuid")
		return nil, constants.ErrNotFound
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to get note by share_uuid")
		return nil, fmt.Errorf("failed to get note by share_uuid: %w", err)
	}

	if parentNoteID.Valid {
		val := uint64(parentNoteID.Int64)
		note.ParentNoteID = &val
	}
	if iconFileID.Valid {
		val := uint64(iconFileID.Int64)
		note.IconFileID = &val
	}
	if shareUUIDResult.Valid {
		note.ShareUUID = &shareUUIDResult.String
	}

	return note, nil
}
