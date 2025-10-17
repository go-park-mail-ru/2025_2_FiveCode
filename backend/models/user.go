package models

import "time"

// User представляет пользователя — используется в ответах API (пароль скрыт).
type User struct {
	ID        uint64    `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}
