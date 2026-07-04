package service

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrNotFound           = errors.New("recipe not found")
	ErrForbidden          = errors.New("forbidden")
	ErrSharedUserNotFound = errors.New("shared user not found")
	ErrDuplicate          = errors.New("duplicate")
)
