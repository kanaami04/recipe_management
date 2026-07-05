package service

import (
	"context"
	"strings"
	"time"

	"recipe-backend/internal/domain"
	"recipe-backend/internal/pkg/invite"
)

// inviteCodeTTL は招待コードの有効期間。
const inviteCodeTTL = 7 * 24 * time.Hour

// defaultGroupName は名前未指定時のグループ名。
const defaultGroupName = "マイグループ"

type ShareGroupService interface {
	// GetMine は userID が所属するグループを返す(未所属なら nil)。
	GetMine(ctx context.Context, userID string) (*domain.ShareGroup, error)
	// Create はグループを作成し、userID を所有者兼メンバーにする。既に所属していれば ErrAlreadyInGroup。
	Create(ctx context.Context, userID, name string) (*domain.ShareGroup, error)
	// Join は招待コードでグループに参加する。既に所属していれば ErrAlreadyInGroup、
	// コードが無効/期限切れなら ErrInviteCodeInvalid。
	Join(ctx context.Context, userID, code string) (*domain.ShareGroup, error)
	// Leave はグループを抜ける。所有者が抜ける場合はグループを解散する。未所属なら ErrNotInGroup。
	Leave(ctx context.Context, userID string) error
	// RemoveMember は所有者が他メンバーを外す。所有者以外の操作は ErrNotGroupOwner。
	RemoveMember(ctx context.Context, ownerID, targetUserID string) error
	// RegenerateInviteCode は所有者が招待コードを再発行する(旧コードを失効)。
	RegenerateInviteCode(ctx context.Context, ownerID string) (*domain.ShareGroup, error)
}

type shareGroupService struct {
	groups domain.ShareGroupRepository
}

func NewShareGroupService(groups domain.ShareGroupRepository) ShareGroupService {
	return &shareGroupService{groups: groups}
}

func (s *shareGroupService) GetMine(ctx context.Context, userID string) (*domain.ShareGroup, error) {
	return s.groups.FindByUserID(ctx, userID)
}

func (s *shareGroupService) Create(ctx context.Context, userID, name string) (*domain.ShareGroup, error) {
	existing, err := s.groups.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAlreadyInGroup
	}
	code, err := invite.Code()
	if err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = defaultGroupName
	}
	group := &domain.ShareGroup{
		Name:                name,
		OwnerID:             userID,
		InviteCode:          code,
		InviteCodeExpiresAt: time.Now().Add(inviteCodeTTL),
	}
	if err := s.groups.Create(ctx, group); err != nil {
		// 別リクエストが同時に参加/作成した場合は user_id の一意制約で弾かれる。
		// 事前チェックを抜けた競合を ErrAlreadyInGroup に正規化する(生 DB エラーで 500 にしない)。
		return nil, s.mapMembershipConflict(ctx, userID, err)
	}
	return s.groups.FindByID(ctx, group.ID)
}

// mapMembershipConflict は、書き込み後に userID が既にどこかのグループへ所属していれば
// ErrAlreadyInGroup を返す(check-then-insert の競合を一意制約で受けたケース)。そうでなければ
// 元のエラーをそのまま返す。
func (s *shareGroupService) mapMembershipConflict(ctx context.Context, userID string, err error) error {
	if existing, ferr := s.groups.FindByUserID(ctx, userID); ferr == nil && existing != nil {
		return ErrAlreadyInGroup
	}
	return err
}

func (s *shareGroupService) Join(ctx context.Context, userID, code string) (*domain.ShareGroup, error) {
	existing, err := s.groups.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAlreadyInGroup
	}
	group, err := s.groups.FindByInviteCode(ctx, strings.TrimSpace(code))
	if err != nil {
		return nil, err
	}
	if group == nil || time.Now().After(group.InviteCodeExpiresAt) {
		return nil, ErrInviteCodeInvalid
	}
	if err := s.groups.AddMember(ctx, group.ID, userID); err != nil {
		return nil, s.mapMembershipConflict(ctx, userID, err)
	}
	return s.groups.FindByID(ctx, group.ID)
}

func (s *shareGroupService) Leave(ctx context.Context, userID string) error {
	group, err := s.groups.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if group == nil {
		return ErrNotInGroup
	}
	// 所有者が抜ける場合はグループを解散する(メンバー行は CASCADE で消える)。
	if group.OwnerID == userID {
		return s.groups.Delete(ctx, group.ID)
	}
	return s.groups.RemoveMember(ctx, group.ID, userID)
}

func (s *shareGroupService) RemoveMember(ctx context.Context, ownerID, targetUserID string) error {
	group, err := s.groups.FindByUserID(ctx, ownerID)
	if err != nil {
		return err
	}
	if group == nil {
		return ErrNotInGroup
	}
	if group.OwnerID != ownerID {
		return ErrNotGroupOwner
	}
	// 所有者自身は外せない(抜けるなら Leave = 解散)。
	if targetUserID == ownerID {
		return ErrForbidden
	}
	return s.groups.RemoveMember(ctx, group.ID, targetUserID)
}

func (s *shareGroupService) RegenerateInviteCode(ctx context.Context, ownerID string) (*domain.ShareGroup, error) {
	group, err := s.groups.FindByUserID(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, ErrNotInGroup
	}
	if group.OwnerID != ownerID {
		return nil, ErrNotGroupOwner
	}
	code, err := invite.Code()
	if err != nil {
		return nil, err
	}
	if err := s.groups.UpdateInviteCode(ctx, group.ID, code, time.Now().Add(inviteCodeTTL)); err != nil {
		return nil, err
	}
	return s.groups.FindByID(ctx, group.ID)
}
