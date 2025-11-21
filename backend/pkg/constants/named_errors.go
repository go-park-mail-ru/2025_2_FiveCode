package constants

import "errors"

var (
	ErrUserExists             = errors.New("user already exists")
	ErrInvalidEmailOrPassword = errors.New("invalid email or password")
	ErrNotFound               = errors.New("not found")
	ErrInvalidSession         = errors.New("invalid session")
	ErrNoAccess               = errors.New("no access")
	ErrInvalidFileType        = errors.New("invalid file type")

	ErrNoteNotFound  = errors.New("note not found")
	ErrAccessDenied  = errors.New("access denied")
	ErrBlockNotFound = errors.New("block not found")
)
