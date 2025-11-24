package usecase

import (
	"context"
	"fmt"

	"backend/notes_service/internal/models"
)

type SharingRepository interface {
	// Collaborators management
	AddCollaborator(ctx context.Context, permission *models.NotePermission) (*models.NotePermission, error)
	GetCollaboratorsByNoteID(ctx context.Context, noteID uint64) ([]*models.NotePermission, error)
	GetCollaboratorByID(ctx context.Context, permissionID uint64) (*models.NotePermission, error)
	UpdateCollaboratorRole(ctx context.Context, permissionID uint64, role models.NoteRole) error
	RemoveCollaborator(ctx context.Context, permissionID uint64) error
	CheckCollaboratorExists(ctx context.Context, noteID, userID uint64) (bool, error)

	// Public access management
	SetPublicAccess(ctx context.Context, noteID uint64, accessLevel *models.NoteRole) error
	GetPublicAccess(ctx context.Context, noteID uint64) (*models.NoteRole, error)

	// Note ownership and access checks
	GetNoteOwnerID(ctx context.Context, noteID uint64) (uint64, error)
	CheckNoteAccess(ctx context.Context, noteID, userID uint64) (*models.NoteAccessInfo, error)
	IsNoteOwner(ctx context.Context, noteID, userID uint64) (bool, error)
	GetUserPermission(ctx context.Context, noteID, userID uint64) (*models.NotePermission, error)
	CanUserShare(ctx context.Context, noteID, userID uint64) (bool, error)

	// is_shared flag management
	UpdateIsSharedFlag(ctx context.Context, noteID uint64, isShared bool) error
}

type NotesRepository interface {
	GetNoteByShareUUID(ctx context.Context, shareUUID string) (*models.Note, error)
}

type SharingUsecase struct {
	sharingRepo SharingRepository
	notesRepo   NotesRepository
}

func NewSharingUsecase(sharingRepo SharingRepository, notesRepo NotesRepository) *SharingUsecase {
	return &SharingUsecase{
		sharingRepo: sharingRepo,
		notesRepo:   notesRepo,
	}
}

func (uc *SharingUsecase) validateNoteOwnership(ctx context.Context, noteID, userID uint64) error {
	isOwner, err := uc.sharingRepo.IsNoteOwner(ctx, noteID, userID)
	if err != nil {
		return fmt.Errorf("failed to check note ownership: %w", err)
	}

	if !isOwner {
		return fmt.Errorf("access denied: user is not the note owner")
	}

	return nil
}

func (uc *SharingUsecase) validateNoteAccess(ctx context.Context, noteID, userID uint64) (*models.NoteAccessInfo, error) {
	accessInfo, err := uc.sharingRepo.CheckNoteAccess(ctx, noteID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check note access: %w", err)
	}

	if !accessInfo.HasAccess {
		return nil, fmt.Errorf("access denied: user does not have access to this note")
	}

	return accessInfo, nil
}

func (uc *SharingUsecase) CheckNoteAccess(ctx context.Context, noteID, userID uint64) (*models.NoteAccessInfo, error) {
	accessInfo, err := uc.sharingRepo.CheckNoteAccess(ctx, noteID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check note access: %w", err)
	}

	return accessInfo, nil
}

// updateIsSharedFlag автоматически обновляет флаг is_shared на основе количества коллабораторов
func (uc *SharingUsecase) updateIsSharedFlag(ctx context.Context, noteID uint64) error {
	// Получаем список коллабораторов
	collaborators, err := uc.sharingRepo.GetCollaboratorsByNoteID(ctx, noteID)
	if err != nil {
		return fmt.Errorf("failed to get collaborators count: %w", err)
	}

	// Если есть хотя бы один коллаборатор - is_shared = true, иначе false
	isShared := len(collaborators) > 0

	if err := uc.sharingRepo.UpdateIsSharedFlag(ctx, noteID, isShared); err != nil {
		return fmt.Errorf("failed to update is_shared flag: %w", err)
	}

	return nil
}

func (uc *SharingUsecase) AddCollaborator(ctx context.Context, noteID, currentUserID, targetUserID uint64, role models.NoteRole) (*models.NotePermission, error) {
	// 1. Проверяем, что текущий пользователь - владелец
	if err := uc.validateNoteOwnership(ctx, noteID, currentUserID); err != nil {
		return nil, err
	}

	// 2. Проверяем, что целевой пользователь не владелец (нельзя добавить владельца как collaborator)
	isOwner, err := uc.sharingRepo.IsNoteOwner(ctx, noteID, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if target user is owner: %w", err)
	}
	if isOwner {
		return nil, fmt.Errorf("cannot add note owner as collaborator")
	}

	// 3. Проверяем, что пользователь еще не имеет доступа
	exists, err := uc.sharingRepo.CheckCollaboratorExists(ctx, noteID, targetUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check collaborator exists: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("user already has access to this note")
	}

	// 4. Создаем разрешение
	permission := &models.NotePermission{
		NoteID:    noteID,
		GrantedBy: currentUserID,
		GrantedTo: targetUserID,
		Role:      role,
	}

	createdPermission, err := uc.sharingRepo.AddCollaborator(ctx, permission)
	if err != nil {
		return nil, fmt.Errorf("failed to add collaborator: %w", err)
	}

	// 5. Обновляем флаг is_shared
	if err := uc.updateIsSharedFlag(ctx, noteID); err != nil {
		return nil, fmt.Errorf("failed to update is_shared flag: %w", err)
	}

	return createdPermission, nil
}

// ============================================
// GetCollaborators - Получить список редакторов
// ============================================

// GetCollaborators возвращает список всех collaborators для заметки
func (uc *SharingUsecase) GetCollaborators(ctx context.Context, noteID, currentUserID uint64) (uint64, []*models.NotePermission, *models.NoteRole, error) {
	// 1. Проверяем доступ к заметке
	_, err := uc.validateNoteAccess(ctx, noteID, currentUserID)
	if err != nil {
		return 0, nil, nil, err
	}

	// 2. Получаем owner_id
	ownerID, err := uc.sharingRepo.GetNoteOwnerID(ctx, noteID)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to get note owner: %w", err)
	}

	// 3. Получаем список collaborators
	permissions, err := uc.sharingRepo.GetCollaboratorsByNoteID(ctx, noteID)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to get collaborators: %w", err)
	}

	// 4. Получаем публичный доступ
	publicAccess, err := uc.sharingRepo.GetPublicAccess(ctx, noteID)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to get public access: %w", err)
	}

	return ownerID, permissions, publicAccess, nil
}

// ============================================
// UpdateCollaboratorRole - Изменить роль редактора
// ============================================

// UpdateCollaboratorRole изменяет роль collaborator
func (uc *SharingUsecase) UpdateCollaboratorRole(ctx context.Context, noteID, currentUserID, permissionID uint64, newRole models.NoteRole) (*models.NotePermission, error) {
	// 1. Проверяем, что текущий пользователь - владелец
	if err := uc.validateNoteOwnership(ctx, noteID, currentUserID); err != nil {
		return nil, err
	}

	// 2. Проверяем, что разрешение существует и принадлежит этой заметке
	permission, err := uc.sharingRepo.GetCollaboratorByID(ctx, permissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get collaborator: %w", err)
	}

	if permission.NoteID != noteID {
		return nil, fmt.Errorf("permission does not belong to this note")
	}

	// 3. Обновляем роль
	if err := uc.sharingRepo.UpdateCollaboratorRole(ctx, permissionID, newRole); err != nil {
		return nil, fmt.Errorf("failed to update collaborator role: %w", err)
	}

	// 4. Получаем обновленное разрешение
	updatedPermission, err := uc.sharingRepo.GetCollaboratorByID(ctx, permissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated collaborator: %w", err)
	}

	return updatedPermission, nil
}

// ============================================
// RemoveCollaborator - Удалить редактора
// ============================================

// RemoveCollaborator удаляет collaborator
func (uc *SharingUsecase) RemoveCollaborator(ctx context.Context, noteID, currentUserID, permissionID uint64) error {
	// 1. Проверяем, что текущий пользователь - владелец
	if err := uc.validateNoteOwnership(ctx, noteID, currentUserID); err != nil {
		return err
	}

	// 2. Проверяем, что разрешение существует и принадлежит этой заметке
	permission, err := uc.sharingRepo.GetCollaboratorByID(ctx, permissionID)
	if err != nil {
		return fmt.Errorf("failed to get collaborator: %w", err)
	}

	if permission.NoteID != noteID {
		return fmt.Errorf("permission does not belong to this note")
	}

	// 3. Удаляем разрешение
	if err := uc.sharingRepo.RemoveCollaborator(ctx, permissionID); err != nil {
		return fmt.Errorf("failed to remove collaborator: %w", err)
	}

	// 4. Обновляем флаг is_shared
	if err := uc.updateIsSharedFlag(ctx, noteID); err != nil {
		return fmt.Errorf("failed to update is_shared flag: %w", err)
	}

	return nil
}

// ============================================
// SetPublicAccess - Установить публичный доступ
// ============================================

// SetPublicAccess устанавливает или отключает публичный доступ к заметке
func (uc *SharingUsecase) SetPublicAccess(ctx context.Context, noteID, currentUserID uint64, accessLevel *models.NoteRole) error {
	// 1. Проверяем, что текущий пользователь - владелец
	if err := uc.validateNoteOwnership(ctx, noteID, currentUserID); err != nil {
		return err
	}

	// 2. Устанавливаем публичный доступ
	if err := uc.sharingRepo.SetPublicAccess(ctx, noteID, accessLevel); err != nil {
		return fmt.Errorf("failed to set public access: %w", err)
	}

	return nil
}

// GetPublicAccess возвращает настройки публичного доступа
func (uc *SharingUsecase) GetPublicAccess(ctx context.Context, noteID, currentUserID uint64) (*models.NoteRole, error) {
	// 1. Проверяем доступ к заметке
	_, err := uc.validateNoteAccess(ctx, noteID, currentUserID)
	if err != nil {
		return nil, err
	}

	// 2. Получаем публичный доступ
	accessLevel, err := uc.sharingRepo.GetPublicAccess(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public access: %w", err)
	}

	return accessLevel, nil
}

func (uc *SharingUsecase) GetSharingSettings(ctx context.Context, noteID, currentUserID uint64) (*models.SharingSettings, error) {
	_, err := uc.validateNoteAccess(ctx, noteID, currentUserID)
	if err != nil {
		return nil, err
	}

	// 2. Получаем owner_id
	ownerID, err := uc.sharingRepo.GetNoteOwnerID(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get note owner: %w", err)
	}

	// 3. Получаем список collaborators
	permissions, err := uc.sharingRepo.GetCollaboratorsByNoteID(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get collaborators: %w", err)
	}

	// 4. Получаем публичный доступ
	publicAccessLevel, err := uc.sharingRepo.GetPublicAccess(ctx, noteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public access: %w", err)
	}

	// 5. Конвертируем permissions в Collaborators
	// Gateway заполнит детали пользователей (username, email, avatar)
	collaborators := make([]models.Collaborator, 0, len(permissions))
	for _, p := range permissions {
		collaborators = append(collaborators, models.Collaborator{
			PermissionID: p.PermissionID,
			UserID:       p.GrantedTo,
			Role:         p.Role,
			GrantedBy:    p.GrantedBy,
			GrantedAt:    p.CreatedAt,
			// Username, Email, AvatarURL заполнит Gateway
		})
	}

	// 6. Формируем Owner (Gateway заполнит детали)
	owner := models.NoteOwner{
		UserID: ownerID,
		// Username, Email, AvatarURL заполнит Gateway
	}

	// 7. Формируем PublicAccess
	publicAccess := models.PublicAccess{
		NoteID:      noteID,
		AccessLevel: publicAccessLevel,
		ShareURL:    fmt.Sprintf("/notes/%d", noteID), // базовый URL
	}

	// 8. Проверяем, является ли текущий пользователь владельцем
	isOwner := (currentUserID == ownerID)

	// 9. Формируем итоговую структуру
	settings := &models.SharingSettings{
		NoteID:        noteID,
		Owner:         owner,
		PublicAccess:  publicAccess,
		Collaborators: collaborators,
		TotalCount:    len(collaborators),
		IsOwner:       isOwner,
	}

	return settings, nil
}

// ActivateAccessByLink создает note_permission при переходе по публичной ссылке
func (uc *SharingUsecase) ActivateAccessByLink(ctx context.Context, shareUUID string, userID uint64) (*models.ActivateAccessResponse, error) {
	// 1. Получаем заметку по share_uuid
	note, err := uc.notesRepo.GetNoteByShareUUID(ctx, shareUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get note by share_uuid: %w", err)
	}

	// 2. Проверяем, что пользователь не является владельцем
	if note.OwnerID == userID {
		// Владелец уже имеет полный доступ, не нужно создавать permission
		return &models.ActivateAccessResponse{
			NoteID:        note.ID,
			AccessGranted: true,
			AccessInfo: models.NoteAccessInfo{
				IsOwner:   true,
				HasAccess: true,
				Role:      models.RoleEditor,
				CanEdit:   true,
			},
		}, nil
	}

	// 3. Проверяем, есть ли уже разрешение для этого пользователя
	existingPermission, err := uc.sharingRepo.GetUserPermission(ctx, note.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing permission: %w", err)
	}

	if existingPermission != nil {
		// Разрешение уже существует, возвращаем текущий доступ
		return &models.ActivateAccessResponse{
			NoteID:        note.ID,
			AccessGranted: true,
			AccessInfo: models.NoteAccessInfo{
				IsOwner:   false,
				HasAccess: true,
				Role:      existingPermission.Role,
				CanEdit:   existingPermission.Role == models.RoleEditor,
			},
		}, nil
	}

	// 4. Получаем публичный уровень доступа
	publicAccessLevel, err := uc.sharingRepo.GetPublicAccess(ctx, note.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public access level: %w", err)
	}

	if publicAccessLevel == nil {
		// Публичный доступ не настроен
		return &models.ActivateAccessResponse{
			NoteID:        note.ID,
			AccessGranted: false,
			AccessInfo: models.NoteAccessInfo{
				IsOwner:   false,
				HasAccess: false,
			},
		}, nil
	}

	// 5. Создаем permission с ролью из public_access_level
	permission := &models.NotePermission{
		NoteID:    note.ID,
		GrantedBy: note.OwnerID, // Доступ предоставлен владельцем через публичную ссылку
		GrantedTo: userID,
		Role:      *publicAccessLevel,
	}

	createdPermission, err := uc.sharingRepo.AddCollaborator(ctx, permission)
	if err != nil {
		return nil, fmt.Errorf("failed to create permission by link: %w", err)
	}

	// 6. Обновляем флаг is_shared
	if err := uc.updateIsSharedFlag(ctx, note.ID); err != nil {
		return nil, fmt.Errorf("failed to update is_shared flag: %w", err)
	}

	// 7. Возвращаем информацию о доступе с note_id
	return &models.ActivateAccessResponse{
		NoteID:        note.ID,
		AccessGranted: true,
		AccessInfo: models.NoteAccessInfo{
			IsOwner:   false,
			HasAccess: true,
			Role:      createdPermission.Role,
			CanEdit:   createdPermission.Role == models.RoleEditor,
		},
	}, nil
}
