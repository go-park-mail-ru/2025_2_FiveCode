package models

type UpdateUserInput struct {
	ID           uint64
	Username     *string
	Password     *string
	AvatarFileID *uint64
}