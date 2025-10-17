package namederrors

import "errors"

var (
	ErrUserExists             = errors.New("user already exists")
	ErrInvalidEmailOrPassword = errors.New("invalid email or password")
	ErrNotFound               = errors.New("not found")
	ErrNoCookie               = errors.New("no cookie")
	ErrInvalidSession         = errors.New("invalid session")
)