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
)

type SharingRepository struct {
	db store.DB
}

func NewSharingRepository(db store.DB) *SharingRepository {
	return &SharingRepository{db: db}
}

func (r *SharingRepository) AddCollaborator(ctx context.Context, permission *models.NotePermission) (*models.NotePermission, error) {
	log := logger.FromContext(ctx)

	query := `
		INSERT INTO note_permission (note_id, granted_by, granted_to, role)
		VALUES ($1, $2, $3, $4)
		RETURNING note_permission_id, created_at, updated_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		permission.NoteID,
		permission.GrantedBy,
		permission.GrantedTo,
		permission.Role,
	).Scan(&permission.PermissionID, &permission.CreatedAt, &permission.UpdatedAt)

	if err != nil {
		log.Error().Err(err).Msg("failed to add collaborator")
		return nil, fmt.Errorf("failed to add collaborator: %w", err)
	}

	log.Info().
		Uint64("permission_id", permission.PermissionID).
		Uint64("note_id", permission.NoteID).
		Uint64("granted_to", permission.GrantedTo).
		Msg("collaborator added successfully")

	return permission, nil
}

func (r *SharingRepository) GetCollaboratorsByNoteID(ctx context.Context, noteID uint64) ([]*models.NotePermission, error) {
	log := logger.FromContext(ctx)

	query := `
		SELECT 
			np.note_permission_id,
			np.note_id,
			np.granted_by,
			np.granted_to,
			np.role,
			np.created_at,
			np.updated_at
		FROM note_permission np
		INNER JOIN note n ON np.note_id = n.id
		WHERE np.note_id = $1 AND n.deleted_at IS NULL
		ORDER BY np.created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, noteID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get collaborators")
		return nil, fmt.Errorf("failed to get collaborators: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close rows")
		}
	}()

	var permissions []*models.NotePermission
	for rows.Next() {
		var p models.NotePermission
		err := rows.Scan(
			&p.PermissionID,
			&p.NoteID,
			&p.GrantedBy,
			&p.GrantedTo,
			&p.Role,
			&p.CreatedAt,
			&p.UpdatedAt,
		)
		if err != nil {
			log.Error().Err(err).Msg("failed to scan permission")
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, &p)
	}

	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("error iterating permissions")
		return nil, fmt.Errorf("error iterating permissions: %w", err)
	}

	return permissions, nil
}

func (r *SharingRepository) GetCollaboratorByID(ctx context.Context, permissionID uint64) (*models.NotePermission, error) {
	log := logger.FromContext(ctx)

	query := `
		SELECT 
			note_permission_id,
			note_id,
			granted_by,
			granted_to,
			role,
			created_at,
			updated_at
		FROM note_permission
		WHERE note_permission_id = $1
	`

	var p models.NotePermission
	err := r.db.QueryRowContext(ctx, query, permissionID).Scan(
		&p.PermissionID,
		&p.NoteID,
		&p.GrantedBy,
		&p.GrantedTo,
		&p.Role,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		log.Warn().Uint64("permission_id", permissionID).Msg("permission not found")
		return nil, constants.ErrNotFound
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to get collaborator")
		return nil, fmt.Errorf("failed to get collaborator: %w", err)
	}

	return &p, nil
}

func (r *SharingRepository) UpdateCollaboratorRole(ctx context.Context, permissionID uint64, role models.NoteRole) error {
	log := logger.FromContext(ctx)

	query := `
		UPDATE note_permission
		SET role = $1, updated_at = NOW()
		WHERE note_permission_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, role, permissionID)
	if err != nil {
		log.Error().Err(err).Msg("failed to update collaborator role")
		return fmt.Errorf("failed to update collaborator role: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("failed to get rows affected")
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		log.Warn().Uint64("permission_id", permissionID).Msg("permission not found for update")
		return constants.ErrNotFound
	}

	log.Info().Uint64("permission_id", permissionID).Str("new_role", string(role)).Msg("collaborator role updated")

	return nil
}

func (r *SharingRepository) RemoveCollaborator(ctx context.Context, permissionID uint64) error {
	log := logger.FromContext(ctx)

	query := `DELETE FROM note_permission WHERE note_permission_id = $1`

	result, err := r.db.ExecContext(ctx, query, permissionID)
	if err != nil {
		log.Error().Err(err).Msg("failed to remove collaborator")
		return fmt.Errorf("failed to remove collaborator: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("failed to get rows affected")
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		log.Warn().Uint64("permission_id", permissionID).Msg("permission not found for deletion")
		return constants.ErrNotFound
	}

	log.Info().Uint64("permission_id", permissionID).Msg("collaborator removed")

	return nil
}

func (r *SharingRepository) CheckCollaboratorExists(ctx context.Context, noteID, userID uint64) (bool, error) {
	log := logger.FromContext(ctx)

	query := `
		SELECT EXISTS(
			SELECT 1 FROM note_permission
			WHERE note_id = $1 AND granted_to = $2
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, noteID, userID).Scan(&exists)
	if err != nil {
		log.Error().Err(err).Msg("failed to check collaborator exists")
		return false, fmt.Errorf("failed to check collaborator exists: %w", err)
	}

	return exists, nil
}

func (r *SharingRepository) SetPublicAccess(ctx context.Context, noteID uint64, accessLevel *models.NoteRole) error {
	log := logger.FromContext(ctx)

	query := `
		UPDATE note
		SET public_access_level = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`

	var level interface{}
	if accessLevel != nil {
		level = *accessLevel
	} else {
		level = nil
	}

	result, err := r.db.ExecContext(ctx, query, level, noteID)
	if err != nil {
		log.Error().Err(err).Msg("failed to set public access")
		return fmt.Errorf("failed to set public access: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("failed to get rows affected")
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		log.Warn().Uint64("note_id", noteID).Msg("note not found for public access update")
		return constants.ErrNotFound
	}

	if accessLevel != nil {
		log.Info().Uint64("note_id", noteID).Str("access_level", string(*accessLevel)).Msg("public access enabled")
	} else {
		log.Info().Uint64("note_id", noteID).Msg("public access disabled")
	}

	return nil
}

func (r *SharingRepository) GetPublicAccess(ctx context.Context, noteID uint64) (*models.NoteRole, string, error) {
	log := logger.FromContext(ctx)

	query := `
        SELECT public_access_level, share_uuid
        FROM note
        WHERE id = $1 AND deleted_at IS NULL
    `

	var accessLevel sql.NullString
	var shareUUID sql.NullString

	err := r.db.QueryRowContext(ctx, query, noteID).Scan(&accessLevel, &shareUUID)
	if errors.Is(err, sql.ErrNoRows) {
		log.Warn().Uint64("note_id", noteID).Msg("note not found")
		return nil, "", constants.ErrNotFound
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to get public access")
		return nil, "", fmt.Errorf("failed to get public access: %w", err)
	}

	var role *models.NoteRole
	if accessLevel.Valid {
		r := models.NoteRole(accessLevel.String)
		role = &r
	}

	uuid := ""
	if shareUUID.Valid {
		uuid = shareUUID.String
	}

	return role, uuid, nil
}

func (r *SharingRepository) GetNoteOwnerID(ctx context.Context, noteID uint64) (uint64, error) {
	log := logger.FromContext(ctx)

	query := `SELECT owner_id FROM note WHERE id = $1 AND deleted_at IS NULL`

	var ownerID uint64
	err := r.db.QueryRowContext(ctx, query, noteID).Scan(&ownerID)
	if errors.Is(err, sql.ErrNoRows) {
		log.Warn().Uint64("note_id", noteID).Msg("note not found")
		return 0, constants.ErrNotFound
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to get note owner")
		return 0, fmt.Errorf("failed to get note owner: %w", err)
	}

	return ownerID, nil
}

func (r *SharingRepository) CheckNoteAccess(ctx context.Context, noteID, userID uint64) (*models.NoteAccessInfo, error) {
	log := logger.FromContext(ctx)

	parentQuery := `
		SELECT parent_note_id 
		FROM note 
		WHERE id = $1 AND deleted_at IS NULL
	`

	var parentNoteID sql.NullInt64
	err := r.db.QueryRowContext(ctx, parentQuery, noteID).Scan(&parentNoteID)
	if errors.Is(err, sql.ErrNoRows) {
		log.Warn().Uint64("note_id", noteID).Msg("note not found for access check")
		return &models.NoteAccessInfo{HasAccess: false}, nil
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to get note for access check")
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	checkNoteID := noteID
	if parentNoteID.Valid {
		checkNoteID = uint64(parentNoteID.Int64)
		log.Info().
			Uint64("sub_note_id", noteID).
			Uint64("parent_note_id", checkNoteID).
			Msg("checking access for sub-note via parent")
	}

	query := `
		SELECT 
			n.owner_id,
			np.role
		FROM note n
		LEFT JOIN note_permission np ON n.id = np.note_id AND np.granted_to = $2
		WHERE n.id = $1 AND n.deleted_at IS NULL
	`

	var ownerID uint64
	var permissionRole sql.NullString

	err = r.db.QueryRowContext(ctx, query, checkNoteID, userID).Scan(&ownerID, &permissionRole)
	if errors.Is(err, sql.ErrNoRows) {
		log.Warn().Uint64("note_id", checkNoteID).Msg("note not found for access check")
		return &models.NoteAccessInfo{HasAccess: false}, nil
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to check note access")
		return nil, fmt.Errorf("failed to check note access: %w", err)
	}

	accessInfo := &models.NoteAccessInfo{}

	if ownerID == userID {
		accessInfo.IsOwner = true
		accessInfo.HasAccess = true
		accessInfo.Role = models.RoleOwner
		accessInfo.CanEdit = true
		return accessInfo, nil
	}

	accessInfo.IsOwner = false

	if permissionRole.Valid {
		accessInfo.HasAccess = true
		accessInfo.Role = models.NoteRole(permissionRole.String)
		accessInfo.CanEdit = (accessInfo.Role == models.RoleEditor)
		return accessInfo, nil
	}

	accessInfo.HasAccess = false
	return accessInfo, nil
}

func (r *SharingRepository) IsNoteOwner(ctx context.Context, noteID, userID uint64) (bool, error) {
	ownerID, err := r.GetNoteOwnerID(ctx, noteID)
	if err != nil {
		return false, err
	}
	return ownerID == userID, nil
}

func (r *SharingRepository) GetUserPermission(ctx context.Context, noteID, userID uint64) (*models.NotePermission, error) {
	log := logger.FromContext(ctx)

	query := `
		SELECT 
			note_permission_id,
			note_id,
			granted_by,
			granted_to,
			role,
			created_at,
			updated_at
		FROM note_permission
		WHERE note_id = $1 AND granted_to = $2
	`

	var p models.NotePermission
	err := r.db.QueryRowContext(ctx, query, noteID, userID).Scan(
		&p.PermissionID,
		&p.NoteID,
		&p.GrantedBy,
		&p.GrantedTo,
		&p.Role,
		&p.CreatedAt,
		&p.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to get user permission")
		return nil, fmt.Errorf("failed to get user permission: %w", err)
	}

	return &p, nil
}

func (r *SharingRepository) CanUserShare(ctx context.Context, noteID, userID uint64) (bool, error) {
	return r.IsNoteOwner(ctx, noteID, userID)
}

func (r *SharingRepository) UpdateIsSharedFlag(ctx context.Context, noteID uint64, isShared bool) error {
	log := logger.FromContext(ctx)

	query := `
		UPDATE note
		SET is_shared = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, isShared, noteID)
	if err != nil {
		log.Error().Err(err).Msg("failed to update is_shared flag")
		return fmt.Errorf("failed to update is_shared flag: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		log.Error().Err(err).Msg("failed to get rows affected")
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		log.Warn().Uint64("note_id", noteID).Msg("note not found for is_shared update")
		return constants.ErrNotFound
	}

	log.Info().Uint64("note_id", noteID).Bool("is_shared", isShared).Msg("is_shared flag updated")
	return nil
}

func (r *SharingRepository) GetParentNoteID(ctx context.Context, noteID uint64) (*uint64, error) {
	log := logger.FromContext(ctx)

	query := `
		SELECT parent_note_id
		FROM note
		WHERE id = $1 AND deleted_at IS NULL
	`

	var parentNoteID sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, noteID).Scan(&parentNoteID)
	if errors.Is(err, sql.ErrNoRows) {
		log.Warn().Uint64("note_id", noteID).Msg("note not found")
		return nil, constants.ErrNotFound
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to get parent note id")
		return nil, fmt.Errorf("failed to get parent note id: %w", err)
	}

	if !parentNoteID.Valid {
		return nil, nil
	}

	parentID := uint64(parentNoteID.Int64)
	return &parentID, nil
}
