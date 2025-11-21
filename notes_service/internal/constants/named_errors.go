package constants

import "errors"

var (
	ErrNotFound               = errors.New("not found")
	ErrNoAccess               = errors.New("no access")
)
