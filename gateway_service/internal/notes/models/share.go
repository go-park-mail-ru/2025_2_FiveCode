package models

import "time"

// NoteRole represents the access level for a note
type NoteRole string

const (
	RoleEditor    NoteRole = "editor"
	RoleCommenter NoteRole = "commenter"
	RoleViewer    NoteRole = "viewer"
)

// ============================================
// Sharing Models
// ============================================

// Collaborator represents a user who has access to a note
type Collaborator struct {
	PermissionID uint64    `json:"permission_id"`
	UserID       uint64    `json:"user_id"`
	Role         NoteRole  `json:"role"`
	GrantedBy    uint64    `json:"granted_by"`
	GrantedAt    time.Time `json:"granted_at"`
}

// CollaboratorResponse represents response after adding/updating collaborator
type CollaboratorResponse struct {
	PermissionID uint64       `json:"permission_id"`
	Collaborator Collaborator `json:"collaborator"`
}

// GetCollaboratorsResponse represents response with all collaborators
type GetCollaboratorsResponse struct {
	NoteID             uint64         `json:"note_id"`
	OwnerID            uint64         `json:"owner_id"`
	Collaborators      []Collaborator `json:"collaborators"`
	PublicAccessLevel  *NoteRole      `json:"public_access_level"`
	TotalCollaborators int            `json:"total_collaborators"`
}

// PublicAccess represents public access settings
type PublicAccess struct {
	NoteID      uint64    `json:"note_id"`
	AccessLevel *NoteRole `json:"access_level"`
	ShareURL    string    `json:"share_url"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PublicAccessResponse represents response for public access operations
type PublicAccessResponse struct {
	NoteID      uint64    `json:"note_id"`
	AccessLevel *NoteRole `json:"access_level"`
	ShareURL    string    `json:"share_url"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SharingSettingsResponse represents complete sharing settings
type SharingSettingsResponse struct {
	NoteID             uint64         `json:"note_id"`
	OwnerID            uint64         `json:"owner_id"`
	PublicAccess       PublicAccess   `json:"public_access"`
	Collaborators      []Collaborator `json:"collaborators"`
	TotalCollaborators int            `json:"total_collaborators"`
	IsOwner            bool           `json:"is_owner"`
}

// NoteAccessInfo represents user's access information
type NoteAccessInfo struct {
	HasAccess  bool     `json:"has_access"`
	Role       NoteRole `json:"role"`
	IsOwner    bool     `json:"is_owner"`
	CanEdit    bool     `json:"can_edit"`
	CanComment bool     `json:"can_comment"`
}

// ActivateAccessResponse represents response after activating access by link
type ActivateAccessResponse struct {
	NoteID        uint64         `json:"note_id"`
	AccessGranted bool           `json:"access_granted"`
	AccessInfo    NoteAccessInfo `json:"access_info"`
}

// ============================================
// Input Models for Repository/Usecase
// ============================================

// AddCollaboratorInput represents input for adding collaborator
type AddCollaboratorInput struct {
	CurrentUserID uint64   `json:"current_user_id"`
	NoteID        uint64   `json:"note_id"`
	UserID        uint64   `json:"user_id"`
	Role          NoteRole `json:"role"`
}

// UpdateCollaboratorRoleInput represents input for updating collaborator role
type UpdateCollaboratorRoleInput struct {
	CurrentUserID uint64   `json:"current_user_id"`
	NoteID        uint64   `json:"note_id"`
	PermissionID  uint64   `json:"permission_id"`
	NewRole       NoteRole `json:"new_role"`
}

// SetPublicAccessInput represents input for setting public access
type SetPublicAccessInput struct {
	CurrentUserID uint64    `json:"current_user_id"`
	NoteID        uint64    `json:"note_id"`
	AccessLevel   *NoteRole `json:"access_level"`
}
