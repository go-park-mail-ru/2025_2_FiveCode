package namederrors

import "errors"

var (
	ErrUserExists             = errors.New("user already exists")
	ErrInvalidEmailOrPassword = errors.New("invalid email or password")
	ErrNotFound               = errors.New("not found")
	ErrNoCookie               = errors.New("no cookie")
	ErrInvalidSession         = errors.New("invalid session")
	ErrUpdateProfile          = errors.New("failed to update profile")
	ErrGetProfile             = errors.New("failed to get profile")
	ErrFileUpload             = errors.New("failed to upload file")
	ErrFileTooLarge           = errors.New("file too large")
	ErrInvalidFileType        = errors.New("invalid file type")
)
