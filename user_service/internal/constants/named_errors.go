package constants

import "errors"

var (
	ErrUserExists             = errors.New("user already exists")
	ErrInvalidEmailOrPassword = errors.New("invalid email or password")
	ErrNotFound               = errors.New("not found")
	ErrInvalidUsername         = errors.New("invalid username")
)
