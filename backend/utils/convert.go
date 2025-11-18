package utils

import (
	"backend/models"

	userPB "backend/gen/go/user"
)

func ProtoUserToModel(p *userPB.User) *models.User {
	if p == nil {
		return nil
	}
	m := &models.User{
		ID:        p.Id,
		Email:     p.Email,
		Username:  p.Username,
		CreatedAt: p.CreatedAt.AsTime(),
	}
	if p.UpdatedAt != nil && p.UpdatedAt.IsValid() {
		updatedTime := p.UpdatedAt.AsTime()
		m.UpdatedAt = &updatedTime
	}
	if p.AvatarFileId != nil {
		m.AvatarFileID = p.AvatarFileId
	}
	return m
}
