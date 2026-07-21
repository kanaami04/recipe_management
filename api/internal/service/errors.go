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
	// ErrInvalidToken は確認/リセットトークンが不正・期限切れのときに返す。
	ErrInvalidToken = errors.New("invalid or expired token")
	// ErrEmailNotVerified はメール未確認のままログインを試みたときに返す。
	ErrEmailNotVerified = errors.New("email not verified")

	// シェアグループ関連。
	ErrAlreadyInGroup    = errors.New("already in a share group")
	ErrNotInGroup        = errors.New("not in a share group")
	ErrNotGroupOwner     = errors.New("not the share group owner")
	ErrInviteCodeInvalid = errors.New("invite code is invalid or expired")
)
