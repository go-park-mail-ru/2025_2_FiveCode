package models

import "time"

// NoteRole represents the access level for a note
type NoteRole string

const (
	RoleEditor    NoteRole = "editor"
	RoleCommenter NoteRole = "commenter"
	RoleViewer    NoteRole = "viewer"
)

// NotePermission represents a user's permission to access a note
type NotePermission struct {
	PermissionID uint64    `json:"permission_id"`
	NoteID       uint64    `json:"note_id"`
	GrantedBy    uint64    `json:"granted_by"`
	GrantedTo    uint64    `json:"granted_to"`
	Role         NoteRole  `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// PublicAccess represents public access settings for a note
type PublicAccess struct {
	NoteID      uint64    `json:"note_id"`
	AccessLevel *NoteRole `json:"access_level"` // NULL означает private
	ShareURL    string    `json:"share_url"`
}

// NoteAccessInfo contains information about user's access to a note
type NoteAccessInfo struct {
	Role      NoteRole `json:"role"`
	CanEdit   bool     `json:"can_edit"`
	IsOwner   bool     `json:"is_owner"`
	HasAccess bool     `json:"has_access"`
}

// ActivateAccessResponse содержит результат активации доступа по ссылке
type ActivateAccessResponse struct {
	NoteID        uint64         `json:"note_id"`
	AccessGranted bool           `json:"access_granted"`
	AccessInfo    NoteAccessInfo `json:"access_info"`
}

// Collaborator represents a user who has access to a note
type Collaborator struct {
	PermissionID uint64    `json:"permission_id"`
	UserID       uint64    `json:"user_id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	AvatarURL    *string   `json:"avatar_url,omitempty"`
	Role         NoteRole  `json:"role"`
	GrantedBy    uint64    `json:"granted_by"`
	GrantedAt    time.Time `json:"granted_at"`
}

// NoteOwner represents the owner of a note
type NoteOwner struct {
	UserID    uint64  `json:"user_id"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// SharingSettings contains all sharing information for a note
type SharingSettings struct {
	NoteID        uint64         `json:"note_id"`
	Owner         NoteOwner      `json:"owner"`
	PublicAccess  PublicAccess   `json:"public_access"`
	Collaborators []Collaborator `json:"collaborators"`
	TotalCount    int            `json:"total_collaborators"`
	IsOwner       bool           `json:"is_owner"` // Может ли текущий пользователь управлять доступами
}
