package service

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNotFound           = errors.New("recipe not found")
	ErrForbidden          = errors.New("forbidden")
	ErrSharedUserNotFound = errors.New("shared user not found")
)
