package constants

import "errors"

var (
	ErrInvalidSession = errors.New("invalid session")
	ErrNotFound       = errors.New("not found")
	ErrInvalidFileType  = errors.New("invalid file type")
)
