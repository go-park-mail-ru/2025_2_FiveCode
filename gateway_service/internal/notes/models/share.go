package models

import "time"

type NoteRole string

const (
	RoleOwner     NoteRole = "owner"
	RoleEditor    NoteRole = "editor"
	RoleCommenter NoteRole = "commenter"
	RoleViewer    NoteRole = "viewer"
)

type Collaborator struct {
	PermissionID uint64    `json:"permission_id"`
	UserID       uint64    `json:"user_id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	AvatarFileID *uint64   `json:"avatar_file_id,omitempty"`
	Role         NoteRole  `json:"role"`
	GrantedBy    uint64    `json:"granted_by"`
	GrantedAt    time.Time `json:"granted_at"`
}

type CollaboratorResponse struct {
	PermissionID uint64       `json:"permission_id"`
	Collaborator Collaborator `json:"collaborator"`
}

type GetCollaboratorsResponse struct {
	NoteID             uint64         `json:"note_id"`
	OwnerID            uint64         `json:"owner_id"`
	OwnerEmail         string         `json:"owner_email"`
	OwnerUsername      string         `json:"owner_username"`
	OwnerAvatarFileID  *uint64        `json:"owner_avatar_file_id,omitempty"`
	Collaborators      []Collaborator `json:"collaborators"`
	PublicAccessLevel  *NoteRole      `json:"public_access_level"`
	TotalCollaborators int            `json:"total_collaborators"`
}

type PublicAccess struct {
	NoteID      uint64    `json:"note_id"`
	AccessLevel *NoteRole `json:"access_level"`
	ShareURL    string    `json:"share_url"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PublicAccessResponse struct {
	NoteID      uint64    `json:"note_id"`
	AccessLevel *NoteRole `json:"access_level"`
	ShareURL    string    `json:"share_url"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SharingSettingsResponse struct {
	NoteID             uint64         `json:"note_id"`
	OwnerID            uint64         `json:"owner_id"`
	OwnerEmail         string         `json:"owner_email"`
	OwnerUsername      string         `json:"owner_username"`
	OwnerAvatarFileID  *uint64        `json:"owner_avatar_file_id,omitempty"`
	PublicAccess       PublicAccess   `json:"public_access"`
	Collaborators      []Collaborator `json:"collaborators"`
	TotalCollaborators int            `json:"total_collaborators"`
	IsOwner            bool           `json:"is_owner"`
}

type NoteAccessInfo struct {
	HasAccess  bool     `json:"has_access"`
	Role       NoteRole `json:"role"`
	IsOwner    bool     `json:"is_owner"`
	CanEdit    bool     `json:"can_edit"`
	CanComment bool     `json:"can_comment"`
}

type ActivateAccessResponse struct {
	NoteID        uint64         `json:"note_id"`
	AccessGranted bool           `json:"access_granted"`
	AccessInfo    NoteAccessInfo `json:"access_info"`
}

type AddCollaboratorInput struct {
	CurrentUserID uint64   `json:"current_user_id"`
	NoteID        uint64   `json:"note_id"`
	Email         string   `json:"email"`
	Role          NoteRole `json:"role"`
}

type UpdateCollaboratorRoleInput struct {
	CurrentUserID uint64   `json:"current_user_id"`
	NoteID        uint64   `json:"note_id"`
	PermissionID  uint64   `json:"permission_id"`
	NewRole       NoteRole `json:"new_role"`
}

type SetPublicAccessInput struct {
	CurrentUserID uint64    `json:"current_user_id"`
	NoteID        uint64    `json:"note_id"`
	AccessLevel   *NoteRole `json:"access_level"`
}

type User struct {
	ID           uint64     `json:"id"`
	Email        string     `json:"email"`
	Username     string     `json:"username"`
	AvatarFileID *uint64    `json:"avatar_file_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}
