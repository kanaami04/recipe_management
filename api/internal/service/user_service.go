package service

import (
	"context"
	"log/slog"
	"strings"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/pkg/id"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
	List(ctx context.Context) ([]domain.User, error)
	// UpdateProfile は username を更新する。他ユーザーと重複したら ErrUserAlreadyExists。
	// メールは本人確認が要るため ChangeEmail で別途扱う。
	UpdateProfile(ctx context.Context, userID, username string) (*domain.User, error)
	// ChangeEmail は現在のパスワードを検証して email を変える。
	// パスワードが違えば ErrIncorrectPassword、他ユーザーと重複したら ErrUserAlreadyExists。
	ChangeEmail(ctx context.Context, userID, email, password string) (*domain.User, error)
	// ChangePassword は現在のパスワードを検証して新しいものに変える。
	// 現在のパスワードが違えば ErrIncorrectPassword。
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error
	// DeleteAccount はユーザーと所有レシピを削除する。
	DeleteAccount(ctx context.Context, userID string) error
	// CreateAvatarUploadURL はプロフィール画像アップロード用の署名付き URL を発行する。
	// 戻り値の key はアップロード後、ConfirmAvatar に渡す。
	CreateAvatarUploadURL(ctx context.Context, userID, contentType string) (uploadURL, key string, err error)
	// ConfirmAvatar はアップロード済みの key を自分のプロフィール画像として確定する。
	// key が自分のもの(CreateAvatarUploadURL で発行した接頭辞)でなければ ErrForbidden。
	ConfirmAvatar(ctx context.Context, userID, key string) (*domain.User, error)
	// DeleteAvatar はプロフィール画像を削除する。未設定でもエラーにしない。
	DeleteAvatar(ctx context.Context, userID string) (*domain.User, error)
}

type userService struct {
	users   domain.UserRepository
	avatars domain.AvatarStorage
}

func NewUserService(users domain.UserRepository, avatars domain.AvatarStorage) UserService {
	return &userService{users: users, avatars: avatars}
}

func (s *userService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return s.users.FindByID(ctx, id)
}

func (s *userService) List(ctx context.Context) ([]domain.User, error) {
	return s.users.FindAll(ctx)
}

func (s *userService) UpdateProfile(ctx context.Context, userID, username string) (*domain.User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrNotFound
	}
	// 変更するときだけ、他ユーザーとの重複を確かめる。
	if username != user.Username {
		if err := s.ensureUsernameFree(ctx, username, userID); err != nil {
			return nil, err
		}
	}
	user.Username = username
	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) ChangeEmail(ctx context.Context, userID, email, password string) (*domain.User, error) {
	user, err := s.authenticate(ctx, userID, password)
	if err != nil {
		return nil, err
	}
	if email != user.Email {
		if err := s.ensureEmailFree(ctx, email, userID); err != nil {
			return nil, err
		}
	}
	user.Email = email
	if err := s.users.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	if _, err := s.authenticate(ctx, userID, currentPassword); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.users.UpdatePassword(ctx, userID, string(hash))
}

func (s *userService) DeleteAccount(ctx context.Context, userID string) error {
	return s.users.Delete(ctx, userID)
}

func (s *userService) CreateAvatarUploadURL(ctx context.Context, userID, contentType string) (string, string, error) {
	key := avatarKeyPrefix(userID) + id.New()
	uploadURL, err := s.avatars.PresignUpload(ctx, key, contentType)
	if err != nil {
		return "", "", err
	}
	return uploadURL, key, nil
}

func (s *userService) ConfirmAvatar(ctx context.Context, userID, key string) (*domain.User, error) {
	// 自分宛てに発行した key かを確かめる(他人の key を confirm させない)。
	if !strings.HasPrefix(key, avatarKeyPrefix(userID)) {
		return nil, ErrForbidden
	}
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrNotFound
	}
	oldKey := user.AvatarKey
	if err := s.users.UpdateAvatarKey(ctx, userID, &key); err != nil {
		return nil, err
	}
	// 差し替え前の画像削除はベストエフォート。失敗しても確定自体は成功させる
	// (孤児オブジェクトが残るだけで、ユーザーから見た状態は正しい)。
	// 同じ key を二重に確定したとき(oldKey == key)は現行画像を消さない。
	if oldKey != nil && *oldKey != key {
		if err := s.avatars.Delete(ctx, *oldKey); err != nil {
			slog.Warn("failed to delete old avatar", "user_id", userID, "key", *oldKey, "error", err)
		}
	}
	user.AvatarKey = &key
	return user, nil
}

func (s *userService) DeleteAvatar(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrNotFound
	}
	if user.AvatarKey != nil {
		key := *user.AvatarKey
		// 先に DB のキーを外し、その後ベストエフォートでオブジェクトを消す。
		// 逆順だと「オブジェクト削除成功 → DB 更新失敗」で avatar_url が 404 を指し、
		// 画像が割れて見える。DB を先に正しくすることでユーザーから見た状態を保つ。
		if err := s.users.UpdateAvatarKey(ctx, userID, nil); err != nil {
			return nil, err
		}
		user.AvatarKey = nil
		if err := s.avatars.Delete(ctx, key); err != nil {
			slog.Warn("failed to delete avatar object", "user_id", userID, "key", key, "error", err)
		}
	}
	return user, nil
}

// avatarKeyPrefix は userID 配下のアバターオブジェクトキーの接頭辞。
// ConfirmAvatar で「自分宛てに発行された key か」を確かめるのに使う。
func avatarKeyPrefix(userID string) string {
	return "avatars/" + userID + "/"
}

// authenticate は userID のユーザーを取得し、現在のパスワードを検証する。
// ChangeEmail / ChangePassword で共通の「本人確認」に使う。
func (s *userService) authenticate(ctx context.Context, userID, password string) (*domain.User, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrNotFound
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, ErrIncorrectPassword
	}
	return user, nil
}

// ensureUsernameFree は username が自分以外に使われていないことを確かめる。
func (s *userService) ensureUsernameFree(ctx context.Context, username, selfID string) error {
	existing, err := s.users.FindByUsername(ctx, username)
	if err != nil {
		return err
	}
	if existing != nil && existing.ID != selfID {
		return ErrUserAlreadyExists
	}
	return nil
}

// ensureEmailFree は email が自分以外に使われていないことを確かめる。
func (s *userService) ensureEmailFree(ctx context.Context, email, selfID string) error {
	existing, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if existing != nil && existing.ID != selfID {
		return ErrUserAlreadyExists
	}
	return nil
}
