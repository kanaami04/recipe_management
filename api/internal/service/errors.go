package service

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrNotFound           = errors.New("recipe not found")
	ErrForbidden          = errors.New("forbidden")
	ErrDuplicate          = errors.New("duplicate")
	ErrIncorrectPassword  = errors.New("incorrect password")
	ErrInvalidURL         = errors.New("invalid url")

	// シェアグループ関連。
	ErrAlreadyInGroup    = errors.New("already in a share group")
	ErrNotInGroup        = errors.New("not in a share group")
	ErrNotGroupOwner     = errors.New("not the share group owner")
	ErrInviteCodeInvalid = errors.New("invite code is invalid or expired")
)
