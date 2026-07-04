package request

import "recipe-backend/internal/apigen"

// 構造体定義は openapi.yaml から生成する。生成型を再エクスポートする。
type (
	UpdateUserRequest            = apigen.UpdateUserRequest
	ChangeEmailRequest           = apigen.ChangeEmailRequest
	ChangePasswordRequest        = apigen.ChangePasswordRequest
	CreateAvatarUploadUrlRequest = apigen.CreateAvatarUploadUrlRequest
	ConfirmAvatarRequest         = apigen.ConfirmAvatarRequest
)
