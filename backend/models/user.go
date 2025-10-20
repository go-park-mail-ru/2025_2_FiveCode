package models

import "time"

// User представляет пользователя — используется в ответах API (пароль скрыт).
type User struct {
	ID            uint64     `json:"id"`
	Email         string     `json:"email"`
	Username      string     `json:"username"`
	AvatarFileID  *uint64    `json:"avatar_file_id,omitempty"`
	Password      string     `json:"-"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
}
